package config

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// stringSlice is a custom type for handling multiple string flags
type stringSlice []string

func (i *stringSlice) String() string {
	return "my string representation"
}

func (i *stringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// Config holds all the configuration for the Harpy application.
type Config struct {
	Targets            []string
	Directory          string
	Concurrency        int
	Timeout            int
	MaxRateLimit       int
	MaxFileSize        int64 // in KB
	NoLimit            bool
	Headers            stringSlice
	Proxy              string
	ProxyFile          string
	OutputFile         string
	Verbose            bool
	JsonOutput         bool
	EnableAST          bool  // Changed from DeepScan to EnableAST
	EnableRegex        bool  // New flag for regex analysis
	InsecureSkipVerify bool
	Silent             bool
	NoColor            bool
	LoadedProxies      []*url.URL
}

// NewConfig creates a new Config object with default values.
func NewConfig() *Config {
	return &Config{
		Targets:            []string{},
		Directory:          "",
		Concurrency:        25,
		Timeout:            10,
		MaxRateLimit:       30,
		MaxFileSize:        10240, // Default 10MB max file size
		NoLimit:            false,
		Headers:            []string{},
		Proxy:              "",
		ProxyFile:          "",
		OutputFile:         "",
		Verbose:            false,
		JsonOutput:         false,
		EnableAST:          true,  // Always enable AST analysis in hybrid mode
		EnableRegex:        true,  // Always enable regex analysis in hybrid mode
		InsecureSkipVerify: false,
		Silent:             false,
		NoColor:            false,
		LoadedProxies:      make([]*url.URL, 0),
	}
}

// showHelp displays custom help message with banner
func (c *Config) showHelp() {
	banner := `
    ██╗  ██╗ █████╗ ██████╗ ██████╗ ██╗   ██╗
    ██║  ██║██╔══██╗██╔══██╗██╔══██╗╚██╗ ██╔╝
    ███████║███████║██████╔╝██████╔╝ ╚████╔╝ 
    ██╔══██║██╔══██║██╔══██╗██╔═══╝   ╚██╔╝  
    ██║  ██║██║  ██║██║  ██║██║        ██║   
    ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝        ╚═╝   
`
	
	fmt.Printf("%s\n", banner)
	fmt.Printf("\t\tEndpoint & Parameter Extraction Tool | v1.0.0 by github.com/rafabd1\n\n")
	
	fmt.Println("DESCRIPTION:")
	fmt.Println("  Harpy is a high-performance web reconnaissance tool for extracting endpoints,")
	fmt.Println("  parameters, and hidden assets from JavaScript files using hybrid analysis.")
	fmt.Println()
	
	fmt.Println("USAGE:")
	fmt.Printf("  harpy [flags]\n\n")
	
	fmt.Println("INPUT OPTIONS:")
	fmt.Println("  -u string    Single target URL or file path")
	fmt.Println("  -f string    File containing list of targets (one per line)")
	fmt.Println("  -d string    Directory to scan for JavaScript/TypeScript/HTML/JSON files")
	fmt.Println()
	
	fmt.Println("NETWORK OPTIONS:")
	fmt.Println("  -c int       Number of concurrent workers (default 25)")
	fmt.Println("  -t int       Request timeout in seconds (default 10)")
	fmt.Println("  -l int       Max requests per second per domain (default 30)")
	fmt.Println("  -H strings   Custom headers (can be used multiple times)")
	fmt.Println("  --proxy      Single proxy server (e.g. http://127.0.0.1:8080)")
	fmt.Println("  -p string    File containing proxy list")
	fmt.Println("  --skip-verify Skip TLS certificate verification")
	fmt.Println()
	
	fmt.Println("OUTPUT OPTIONS:")
	fmt.Println("  -o string    Output file path (default: stdout)")
	fmt.Println("  --json       Output results in JSON format")
	fmt.Println("  -v           Enable verbose output")
	fmt.Println("  --silent     Suppress all output except findings")
	fmt.Println("  --no-color   Disable colored output")
	fmt.Println()
	
	fmt.Println("FILE OPTIONS:")
	fmt.Println("  --max-file-size int  Maximum file size to process in KB (default 10240)")
	fmt.Println("  --no-limit           Disable file size restrictions")
	fmt.Println()
	
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Extract endpoints from a single file")
	fmt.Println("  harpy -u https://example.com/app.js")
	fmt.Println()
	fmt.Println("  # Scan local directory with JSON output")
	fmt.Println("  harpy -d ./js-files --json -o results.json")
	fmt.Println()
	fmt.Println("  # Scan from list with custom headers")
	fmt.Println("  harpy -f targets.txt -H \"User-Agent: Harpy/1.0\"")
	fmt.Println()
	fmt.Println("  # Pipe from other tools")
	fmt.Println("  subfinder -d target.com | httpx | harpy")
	fmt.Println()
	
	fmt.Println("NOTE:")
	fmt.Println("  Harpy uses intelligent hybrid analysis (Regex + AST) with automatic")
	fmt.Println("  fallbacks for optimal results. No manual mode configuration needed.")
}

// Parse populates the Config struct from command-line flags and input sources.
func (c *Config) Parse() error {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	
	// Set custom usage function
	fs.Usage = func() {
		c.showHelp()
	}

	var singleTarget string
	fs.StringVar(&singleTarget, "u", "", "A single target URL or file path.")

	var targetFile string
	fs.StringVar(&targetFile, "f", "", "A file containing a list of targets (URLs or file paths).")

	fs.StringVar(&c.Directory, "d", "", "Path to a local directory to scan for .js and .ts files.")

	fs.IntVar(&c.Concurrency, "c", 25, "Number of concurrent workers.")
	fs.IntVar(&c.Timeout, "t", 10, "Request timeout in seconds.")
	fs.IntVar(&c.MaxRateLimit, "l", 30, "Maximum requests per second per domain in auto-adjustment mode.")
	fs.Int64Var(&c.MaxFileSize, "max-file-size", 10240, "Maximum file size to process in KB.")
	fs.BoolVar(&c.NoLimit, "no-limit", false, "Disable file size limit.")

	fs.StringVar(&c.Proxy, "proxy", "", "A single proxy server (e.g. http://127.0.0.1:8080).")
	fs.StringVar(&c.ProxyFile, "p", "", "File containing a list of proxies (http/https/socks5).")

	fs.StringVar(&c.OutputFile, "o", "", "File to write output to.")
	fs.Var(&c.Headers, "H", "Custom header to include in all requests (can be used multiple times).")

	fs.BoolVar(&c.Verbose, "v", false, "Enable verbose output for debugging.")
	fs.BoolVar(&c.JsonOutput, "json", false, "Enable JSON output format.")
	fs.BoolVar(&c.InsecureSkipVerify, "skip-verify", false, "Skip TLS certificate verification.")
	fs.BoolVar(&c.Silent, "silent", false, "Suppress all output except for findings.")
	fs.BoolVar(&c.NoColor, "no-color", false, "Disable colorized output.")

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
				line := strings.TrimSpace(scanner.Text())
				if line != "" { // Filter out empty lines
					c.Targets = append(c.Targets, line)
				}
			}
		}
	}

	if c.ProxyFile != "" && c.Proxy != "" {
		return fmt.Errorf("cannot use both -proxy and -p flags at the same time")
	}

	// Show help if no input provided
	if len(c.Targets) == 0 && c.Directory == "" {
		c.showHelp()
		os.Exit(0)
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
		line := strings.TrimSpace(scanner.Text())
		if line != "" { // Filter out empty lines
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
} 