package networking

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"DepScout/internal/utils"
)

// Client is a wrapper around an HTTP client with connection pooling and proxy support.
type Client struct {
	logger  *utils.Logger
	client  *http.Client
	headers []string
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
func NewClient(
	logger *utils.Logger,
	timeout int,
	insecureSkipVerify bool,
	headers []string,
	loadedProxies []*url.URL,
) (*Client, error) {

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureSkipVerify,
		},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	// Setup proxy rotation if proxies are available
	if len(loadedProxies) > 0 {
		var proxyCounter uint64
		transport.Proxy = func(req *http.Request) (*url.URL, error) {
			// Use atomic counter for thread-safe round-robin
			count := atomic.AddUint64(&proxyCounter, 1) - 1
			proxy := loadedProxies[count%uint64(len(loadedProxies))]
			return proxy, nil
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(timeout) * time.Second,
	}

	return &Client{
		logger:  logger,
		client:  client,
		headers: headers,
	}, nil
}

// Do executes an HTTP request.
func (c *Client) Do(reqData RequestData) *ResponseData {
	req, err := http.NewRequestWithContext(reqData.Ctx, reqData.Method, reqData.URL, nil)
	if err != nil {
		return &ResponseData{Error: err}
	}

	// Prepare headers
	// Start with default User-Agent
	req.Header.Set("User-Agent", "DepScout/1.0")

	// Apply headers from config
	for _, h := range c.headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			req.Header.Set(key, value)
		}
	}

	// Apply request-specific headers, allowing overrides
	for key, value := range reqData.Headers {
		req.Header.Set(key, value)
	}

	c.logger.Debugf("Requesting %s", req.URL.String())

	resp, err := c.client.Do(req)
	if err != nil {
		return &ResponseData{Error: err}
	}

	return &ResponseData{
		Response:   resp,
		StatusCode: resp.StatusCode,
		Error:      nil,
	}
}

