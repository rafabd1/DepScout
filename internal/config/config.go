package config

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

// Config holds all the configuration for the application.
type Config struct {
	Targets            []string
	Directory          string
	Concurrency        int
	Timeout            int
	ProxyFile          string
	OutputFile         string
	Verbose            bool
	JsonOutput         bool
	DeepScan           bool
	InsecureSkipVerify bool
	ParsedProxies      []ProxyEntry // Holds parsed proxy info
}

// ProxyEntry is a struct to hold the parsed proxy URL and its type
type ProxyEntry struct {
	URL  string
	Type string // "http", "https", "socks5"
}

// NewConfig creates a new Config object with default values.
func NewConfig() *Config {
	return &Config{
		Targets:            []string{},
		Directory:          "",
		Concurrency:        25,
		Timeout:            10,
		ProxyFile:          "",
		OutputFile:         "",
		Verbose:            false,
		JsonOutput:         false,
		DeepScan:           false,
		InsecureSkipVerify: false,
		ParsedProxies:      []ProxyEntry{},
	}
}

// Parse populates the Config struct from command-line flags and input sources.
func (c *Config) Parse() error {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	var singleTarget string
	fs.StringVar(&singleTarget, "u", "", "A single target URL or file path.")
	fs.StringVar(&singleTarget, "url", "", "A single target URL or file path.")

	var targetFile string
	fs.StringVar(&targetFile, "f", "", "A file containing a list of targets (URLs or file paths).")
	fs.StringVar(&targetFile, "file", "", "A file containing a list of targets (URLs or file paths).")

	fs.StringVar(&c.Directory, "d", "", "Path to a local directory to scan for .js and .ts files.")
	fs.StringVar(&c.Directory, "directory", "", "Path to a local directory to scan for .js and .ts files.")

	fs.IntVar(&c.Concurrency, "c", 25, "Number of concurrent workers.")
	fs.IntVar(&c.Concurrency, "concurrency", 25, "Number of concurrent workers.")
	fs.IntVar(&c.Timeout, "t", 10, "Request timeout in seconds.")
	fs.IntVar(&c.Timeout, "timeout", 10, "Request timeout in seconds.")
	fs.StringVar(&c.ProxyFile, "p", "", "File containing a list of proxies (http/https/socks5).")
	fs.StringVar(&c.ProxyFile, "proxy-file", "", "File containing a list of proxies (http/https/socks5).")
	fs.StringVar(&c.OutputFile, "o", "", "File to write output to.")
	fs.StringVar(&c.OutputFile, "output", "", "File to write output to.")

	fs.BoolVar(&c.Verbose, "v", false, "Enable verbose output.")
	fs.BoolVar(&c.Verbose, "verbose", false, "Enable verbose output.")
	fs.BoolVar(&c.JsonOutput, "json", false, "Enable JSON output format.")
	fs.BoolVar(&c.DeepScan, "deep-scan", false, "Enable deep scan using AST parsing (slower).")
	fs.BoolVar(&c.InsecureSkipVerify, "skip-verify", false, "Skip TLS certificate verification.")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	// Load targets from single URL/file, file list, or stdin
	if singleTarget != "" {
		c.Targets = append(c.Targets, singleTarget)
	}
	if targetFile != "" {
		lines, err := readLines(targetFile)
		if err != nil {
			return fmt.Errorf("error reading target file: %w", err)
		}
		c.Targets = append(c.Targets, lines...)
	}

	// Load targets from stdin if no other input is provided
	if len(c.Targets) == 0 && c.Directory == "" {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				c.Targets = append(c.Targets, strings.TrimSpace(scanner.Text()))
			}
		}
	}
	
	if c.ProxyFile != "" {
		// Proxy parsing logic would go here if needed again
	}

	if len(c.Targets) == 0 && c.Directory == "" {
		return fmt.Errorf("no targets provided. Use -u, -f, -d, or pipe from stdin")
	}

	return nil
}

// readLines reads a file and returns its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
} 