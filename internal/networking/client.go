package networking

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"DepScout/internal/config"
	"DepScout/internal/utils"
)

// Client is a wrapper around an HTTP client with connection pooling and proxy support.
type Client struct {
	config     *config.Config
	logger     *utils.Logger
	httpClient *http.Client
	proxyPool  *ProxyPool
}

// RequestData holds data for an HTTP request.
type RequestData struct {
	URL     string
	Method  string
	Body    io.Reader
	Headers map[string]string
	Ctx     context.Context
}

// ResponseData holds data from an HTTP response.
type ResponseData struct {
	Response   *http.Response
	StatusCode int
	Error      error
}

// NewClient creates a new custom HTTP client.
func NewClient(cfg *config.Config, logger *utils.Logger) (*Client, error) {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(cfg.Timeout) * time.Second,
	}

	var proxyPool *ProxyPool
	if cfg.ProxyFile != "" {
		// This logic needs to be re-implemented if proxies are a priority.
		// For now, it's disabled to simplify.
		logger.Warnf("Proxy file specified, but proxy logic is currently simplified. Proxies will not be used.")
	}

	return &Client{
		config:     cfg,
		logger:     logger,
		httpClient: httpClient,
		proxyPool:  proxyPool,
	}, nil
}

// Do performs an HTTP request.
func (c *Client) Do(reqData RequestData) ResponseData {
	req, err := http.NewRequestWithContext(reqData.Ctx, reqData.Method, reqData.URL, reqData.Body)
	if err != nil {
		return ResponseData{Error: err}
	}
	req.Header.Set("User-Agent", "DepScout/1.0") // Set a default user agent

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ResponseData{Error: err}
	}

	return ResponseData{
		Response:   resp,
		StatusCode: resp.StatusCode,
		Error:      nil,
	}
}

// ProxyPool manages a pool of proxy URLs.
type ProxyPool struct {
	proxies []string
	next    uint32
	mu      sync.Mutex
}

// NewProxyPool creates a new ProxyPool.
func NewProxyPool(proxies []string) *ProxyPool {
	return &ProxyPool{
		proxies: proxies,
	}
}

// GetNextProxy cycles through the proxy list.
func (p *ProxyPool) GetNextProxy() (string, error) {
	if len(p.proxies) == 0 {
		return "", fmt.Errorf("proxy pool is empty")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	proxy := p.proxies[p.next%uint32(len(p.proxies))]
	p.next++
	return proxy, nil
}

