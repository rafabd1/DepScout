package utils

import (
	"net/url"
	"strings"
)

// cleanURL remove query parameters and fragments from a URL string.
func cleanURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL // Return original if parsing fails
	}
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

// GetBaseURL returns the scheme and host part of a URL.
func GetBaseURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return u.Scheme + "://" + u.Host, nil
}

// IsSameDomain checks if two URLs belong to the same domain.
func IsSameDomain(url1, url2 string) bool {
	host1, err := GetHost(url1)
	if err != nil {
		return false
	}
	host2, err := GetHost(url2)
	if err != nil {
		return false
	}
	return host1 == host2
}

// GetHost extracts the host from a URL.
func GetHost(rawURL string) (string, error) {
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "http://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return u.Hostname(), nil
} 