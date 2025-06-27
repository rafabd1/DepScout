package networking

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// ParseProxyURL parses a proxy string into a URL object.
// It supports formats like:
// - http://user:pass@host:port
// - host:port
// - host:port:user:pass
func ParseProxyURL(proxyStr string) (*url.URL, error) {
	if !strings.Contains(proxyStr, "://") {
		parts := strings.Split(proxyStr, ":")
		if len(parts) == 2 { // host:port
			proxyStr = fmt.Sprintf("http://%s", proxyStr)
		} else if len(parts) == 4 { // host:port:user:pass
			proxyStr = fmt.Sprintf("http://%s:%s@%s:%s", parts[2], parts[3], parts[0], parts[1])
		} else {
			return nil, fmt.Errorf("invalid proxy format: %s", proxyStr)
		}
	}

	proxyURL, err := url.Parse(proxyStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse proxy URL: %w", err)
	}

	return proxyURL, nil
}

// LoadProxiesFromFile reads a list of proxy strings from a file.
func LoadProxiesFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open proxy file: %w", err)
	}
	defer file.Close()

	var proxies []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			proxies = append(proxies, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading proxy file: %w", err)
	}

	return proxies, nil
}

// CheckProxies tests a list of proxy URLs concurrently and returns the live ones.
func CheckProxies(proxies []*url.URL, timeout int, noColor bool) []*url.URL {
	// Color constants for better styling
	colorCyan := "\033[36m"
	colorGreen := "\033[32m"
	colorReset := "\033[0m"
	
	// Disable colors if requested
	if noColor {
		colorCyan = ""
		colorGreen = ""
		colorReset = ""
	}

	fmt.Printf("%s[INFO]%s Checking proxies... %s⚡%s\n", colorCyan, colorReset, colorCyan, colorReset)

	var liveProxies []*url.URL
	var wg sync.WaitGroup
	var mu sync.Mutex

	// A simple URL to test connectivity
	testURL := "https://www.google.com/gen_204"

	for _, proxy := range proxies {
		wg.Add(1)
		go func(p *url.URL) {
			defer wg.Done()

			transport := &http.Transport{
				Proxy: http.ProxyURL(p),
			}
			client := &http.Client{
				Transport: transport,
				Timeout:   time.Duration(timeout) * time.Second,
			}

			resp, err := client.Get(testURL)
			if err == nil && resp.StatusCode == http.StatusNoContent {
				mu.Lock()
				liveProxies = append(liveProxies, p)
				mu.Unlock()
			}
		}(proxy)
	}

	wg.Wait()

	fmt.Printf("%s[INFO]%s Found %s%d%s live proxies ready for rotation %s🔄%s\n", 
		colorGreen, colorReset, colorGreen, len(liveProxies), colorReset, colorGreen, colorReset)
	
	return liveProxies
} 