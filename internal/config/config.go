package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

// Config struct holds all the configuration for the application.
// It is populated by the command-line flags.
type Config struct {
	Targets            []string
	File               string
	URL                string
	Concurrency        int
	MaxRetries         int
	RequestTimeout     time.Duration
	Silent             bool
	Verbose            bool
	NoColor            bool
	UserAgent          string
	CustomHeaders      []string
	MaxFileSize        int64
	Proxy              string
	ProxyFile          string
	Output             string
	OutputFormat       string // "json" ou "text"
	ParsedProxies      []ProxyEntry
	InsecureSkipVerify bool
}

// ProxyEntry representa um proxy com seu URL.
type ProxyEntry struct {
	URL string
}

// NewDefaultConfig creates a new config with default values.
func NewDefaultConfig() *Config {
	return &Config{
		Targets:            []string{},
		File:               "",
		URL:                "",
		Concurrency:        25,
		RequestTimeout:     10 * time.Second,
		MaxRetries:         2,
		Silent:             false,
		Verbose:            false,
		NoColor:            false,
		UserAgent:          "DepScout - Dependency Confusion Scanner",
		CustomHeaders:      []string{},
		MaxFileSize:        1024 * 1024, // Default 1MB limit
		Proxy:              "",
		ProxyFile:          "",
		Output:             "",
		OutputFormat:       "text",
		ParsedProxies:      []ProxyEntry{},
		InsecureSkipVerify: false,
	}
}

// Validate checks the configuration for any invalid values.
func (c *Config) Validate() error {
	if c.RequestTimeout <= 0 {
		return fmt.Errorf("requestTimeout must be positive")
	}
	if c.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be positive")
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("maxRetries cannot be negative")
	}
	if c.OutputFormat != "text" && c.OutputFormat != "json" {
		return fmt.Errorf("outputFormat must be 'text' or 'json'")
	}
	return nil
}

// ParseAndValidateProxies processa a entrada de proxy (string ou arquivo) e a valida.
func (c *Config) ParseAndValidateProxies() error {
	if c.Proxy == "" && c.ProxyFile == "" {
		return nil
	}

	rawProxies := []string{}
	if c.Proxy != "" {
		rawProxies = append(rawProxies, c.Proxy)
	}

	if c.ProxyFile != "" {
		fileContent, err := os.ReadFile(c.ProxyFile)
		if err != nil {
			return fmt.Errorf("failed to read proxy file %s: %w", c.ProxyFile, err)
		}
		lines := strings.Split(strings.TrimSpace(string(fileContent)), "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				rawProxies = append(rawProxies, trimmed)
			}
		}
	}

	for _, proxyStr := range rawProxies {
		// Validação simples de URL
		_, err := url.Parse(proxyStr)
		if err != nil {
			return fmt.Errorf("invalid proxy URL format '%s': %w", proxyStr, err)
		}
		c.ParsedProxies = append(c.ParsedProxies, ProxyEntry{URL: proxyStr})
	}

	if len(c.ParsedProxies) > 0 && !c.Silent {
		fmt.Printf("[INFO] Loaded %d proxies.\n", len(c.ParsedProxies))
	}

	return nil
} 