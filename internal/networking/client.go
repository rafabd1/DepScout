package networking

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	// "net/http/httputil" // Removido pois não está sendo usado
	"net/url"
	"strings"
	"sync"
	"time"

	"DepScout/internal/config"
	"DepScout/internal/utils"
)

// RequestData agrupa os dados para uma única requisição HTTP.
type RequestData struct {
	URL    string
	Method string
	Ctx    context.Context
}

// ResponseData agrupa os dados de uma resposta HTTP.
type ResponseData struct {
	Response   *http.Response
	Body       []byte
	Error      error
	StatusCode int
}

// Client gerencia requisições HTTP, incluindo headers, retentativas e proxies.
type Client struct {
	httpClient *http.Client
	config     *config.Config
	logger     utils.Logger
	userAgent  string
	parsedProxies    []config.ProxyEntry
	proxyLock        sync.Mutex
	domainProxyIndex map[string]int
	defaultTransport *http.Transport
	currentProxyIndex int // For global round-robin if needed, not primary for domain-specific
	// proxyMutex        sync.Mutex // proxyLock é usado para domainProxyIndex
}

// ClientRequestData struct encapsulates all necessary data for making a request.
type ClientRequestData struct {
	URL            string
	Method         string
	Body           string
	CustomHeaders  http.Header // Alterado para http.Header para facilitar o uso
	RequestHeaders http.Header // Renomeado de CustomHeaders para clareza, usado para headers específicos da requisição
	Ctx            context.Context
}

// ClientResponseData struct holds the outcome of an HTTP request.
type ClientResponseData struct {
	Response    *http.Response
	Body        []byte
	RespHeaders http.Header
	Error       error
	StatusCode  int
}

// NewClient cria um novo cliente HTTP.
func NewClient(cfg *config.Config, logger utils.Logger) (*Client, error) {
	baseTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		},
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   true,
	}

	c := &Client{
		config:           cfg,
		logger:           logger,
		userAgent:        cfg.UserAgent,
		parsedProxies:    cfg.ParsedProxies,
		domainProxyIndex: make(map[string]int),
		defaultTransport: baseTransport,
		currentProxyIndex: 0,
	}

	c.httpClient = &http.Client{
		Transport: c.defaultTransport,
		Timeout:   cfg.RequestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Always return the last response (e.g. the 302 itself)
		},
	}

	return c, nil
}

// getProxyForDomain selects a proxy for a given target domain using round-robin per domain.
func (c *Client) getProxyForDomain(targetDomain string) *url.URL {
	c.proxyLock.Lock()
	defer c.proxyLock.Unlock()

	if len(c.parsedProxies) == 0 {
		return nil
	}

	currentIndex, exists := c.domainProxyIndex[targetDomain]
	if !exists {
		currentIndex = 0
	} else {
		currentIndex = (currentIndex + 1) % len(c.parsedProxies)
	}
	c.domainProxyIndex[targetDomain] = currentIndex

	selectedProxyEntry := c.parsedProxies[currentIndex]
	proxyURL, err := url.Parse(selectedProxyEntry.URL)
	if err != nil {
		c.logger.Warnf("Failed to parse stored proxy URL '%s': %v. Skipping proxy.", selectedProxyEntry.URL, err)
		return nil
	}

	if c.config.Verbose {
		c.logger.Debugf("Selected proxy '%s' for target domain '%s' (Index: %d)", proxyURL.String(), targetDomain, currentIndex)
	}
	return proxyURL
}

// Do executa uma requisição HTTP.
func (c *Client) Do(reqData RequestData) ResponseData {
	req, err := http.NewRequestWithContext(reqData.Ctx, reqData.Method, reqData.URL, nil)
	if err != nil {
		return ResponseData{Error: fmt.Errorf("failed to create request: %w", err)}
	}

	// Adicionar headers
	req.Header.Set("User-Agent", c.config.UserAgent)
	for _, h := range c.config.CustomHeaders {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ResponseData{Error: err}
	}

	return ResponseData{Response: resp, StatusCode: resp.StatusCode}
}

// PerformRequest executes an HTTP request based on the provided ClientRequestData.
func (c *Client) PerformRequest(reqData ClientRequestData) ClientResponseData {
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		req, errBuildReq := http.NewRequestWithContext(reqData.Ctx, reqData.Method, reqData.URL, strings.NewReader(reqData.Body))
		if errBuildReq != nil {
			return ClientResponseData{Error: fmt.Errorf("failed to build request for %s: %w", reqData.URL, errBuildReq)}
		}

		req.Header.Set("User-Agent", c.userAgent)

		// Aplicar headers específicos da requisição (reqData.RequestHeaders)
		if reqData.RequestHeaders != nil {
			for key, values := range reqData.RequestHeaders {
				for _, value := range values {
					req.Header.Add(key, value) // Use Add para suportar múltiplos valores para o mesmo header
				}
			}
		}

		// Aplicar headers customizados globais (c.config.CustomHeaders)
		// Estes são aplicados apenas se não foram definidos pelos headers específicos da requisição.
		for _, headerStr := range c.config.CustomHeaders {
			parts := strings.SplitN(headerStr, ":", 2)
			if len(parts) == 2 {
				headerName := strings.TrimSpace(parts[0])
				headerValue := strings.TrimSpace(parts[1])
				if req.Header.Get(headerName) == "" { // Só adiciona se não foi definido por reqData.RequestHeaders
					req.Header.Set(headerName, headerValue)
				}
			}
		}

		// NOVO: Forçar o header Host se ele for especificado nos headers da requisição.
		// Isso é crucial para virtual hosting e testes como os do Varnish.
		if hostHeaderValue := req.Header.Get("Host"); hostHeaderValue != "" {
			req.Host = hostHeaderValue
		}

		// Determinar o cliente HTTP a ser usado (com ou sem proxy)
		currentHttpClient := c.httpClient
		targetHost := req.URL.Hostname() // Obter o host do URL da requisição

		if len(c.parsedProxies) > 0 {
			selectedProxyURL := c.getProxyForDomain(targetHost)
			if selectedProxyURL != nil {
				proxiedTransport := c.defaultTransport.Clone()
				proxiedTransport.Proxy = http.ProxyURL(selectedProxyURL)
				currentHttpClient = &http.Client{
					Transport:     proxiedTransport,
					Timeout:       c.config.RequestTimeout,
					CheckRedirect: c.httpClient.CheckRedirect,
				}
				if c.config.Verbose {
					c.logger.Debugf("Using proxy %s for request to %s (target host: %s)", selectedProxyURL.String(), reqData.URL, targetHost)
				}
			} else {
				if c.config.Verbose {
					c.logger.Debugf("No proxy selected by getProxyForDomain for %s (target host: %s), using direct connection.", reqData.URL, targetHost)
				}
			}
		}

		if c.config.Verbose {
			c.logger.Debugf("[Client Attempt: %d] Sending %s to %s with headers: %v", attempt+1, reqData.Method, reqData.URL, req.Header)
			if reqData.Body != "" {
				c.logger.Debugf("[Client Attempt: %d] Request body: %s", attempt+1, reqData.Body)
			}
		}

		resp, err := currentHttpClient.Do(req)
		if err != nil {
			// Check if the error indicates a rate-limiting issue from a proxy/CDN that doesn't return a proper 429 response.
			if strings.Contains(strings.ToLower(err.Error()), "too many requests") {
				if c.config.Verbose { // Only log this for -vv
					c.logger.Debugf("Request to %s failed but error indicates rate limiting ('Too Many Requests'). Treating as a 429 response.", reqData.URL)
				}
				// Create a mock 429 response to propagate the rate-limiting signal
				return ClientResponseData{
					Response: &http.Response{
						StatusCode: http.StatusTooManyRequests,
						Status:     "429 Too Many Requests (Inferred from error)",
						Header:     make(http.Header),
						Request:    req,
					},
					StatusCode: http.StatusTooManyRequests,
					Error:      nil, // Clear the original error as we are now handling it as a 429 response
				}
			}

			lastErr = fmt.Errorf("failed on attempt %d for %s: %w", attempt+1, reqData.URL, err)
			if reqData.Ctx.Err() == context.DeadlineExceeded {
				c.logger.Debugf("Request to %s timed out (attempt %d/%d)", reqData.URL, attempt+1, c.config.MaxRetries+1)
			} else if reqData.Ctx.Err() == context.Canceled {
				c.logger.Debugf("Request to %s canceled by context (attempt %d/%d)", reqData.URL, attempt+1, c.config.MaxRetries+1)
			} else if strings.Contains(err.Error(), "dial tcp") && strings.Contains(err.Error(), "timeout") {
				c.logger.Debugf("Request to %s failed with TCP dial timeout (attempt %d/%d): %v", reqData.URL, attempt+1, c.config.MaxRetries+1, err)
			}
			if attempt == c.config.MaxRetries {
				return ClientResponseData{Error: lastErr}
			}
			time.Sleep(1 * time.Second)
			continue
		}

		// Correct handling of response body
		body, errRead := io.ReadAll(resp.Body)
		resp.Body.Close() // Close the body immediately after reading

		if errRead != nil {
			lastErr = fmt.Errorf("failed to read body for %s: %w", reqData.URL, errRead)
			if attempt == c.config.MaxRetries {
				return ClientResponseData{Error: lastErr} // Return error after last retry
			}
			time.Sleep(1 * time.Second) // Wait before next retry
			continue                    // Move to the next attempt
		}

		// Success case
		if c.config.Verbose {
			c.logger.Debugf("Request to %s successful. Status: %s. Body size: %d", reqData.URL, resp.Status, len(body))
		}
		return ClientResponseData{
			Response:    resp,
			Body:        body,
			RespHeaders: resp.Header,
			StatusCode:  resp.StatusCode,
			Error:       nil, // Clear any previous error
		}
	}
	return ClientResponseData{Error: lastErr}
}

