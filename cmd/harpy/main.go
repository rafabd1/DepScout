package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rafabd1/Harpy/internal/config"
	"github.com/rafabd1/Harpy/internal/core"
	"github.com/rafabd1/Harpy/internal/networking"
	"github.com/rafabd1/Harpy/internal/output"
	"github.com/rafabd1/Harpy/internal/report"
	"github.com/rafabd1/Harpy/internal/utils"
)

var version = "1.0.0" // Harpy version

const (
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m" 
	colorReset  = "\033[0m"
)

func main() {
	cfg := config.NewConfig()
	if err := cfg.Parse(); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Use -h or --help for usage.")
		os.Exit(1)
	}

	if !cfg.Silent {
		banner := `
    ██╗  ██╗ █████╗ ██████╗ ██████╗ ██╗   ██╗
    ██║  ██║██╔══██╗██╔══██╗██╔══██╗╚██╗ ██╔╝
    ███████║███████║██████╔╝██████╔╝ ╚████╔╝ 
    ██╔══██║██╔══██║██╔══██╗██╔═══╝   ╚██╔╝  
    ██║  ██║██║  ██║██║  ██║██║        ██║   
    ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝        ╚═╝   `
		
		author := "github.com/rafabd1"
		
		if !cfg.NoColor {
			fmt.Printf("%s%s%s\n", colorCyan, banner, colorReset)
			fmt.Printf("%s\t\tEndpoint & Parameter Extraction Tool | v%s by %s%s\n\n", colorGreen, version, author, colorReset)
		} else {
			fmt.Printf("%s\n", banner)
			fmt.Printf("\t\tEndpoint & Parameter Extraction Tool | v%s by %s\n\n", version, author)
		}
	}

	// --- Proxy Setup ---
	var proxyStrings []string
	if cfg.Proxy != "" {
		proxyStrings = append(proxyStrings, cfg.Proxy)
	}
	if cfg.ProxyFile != "" {
		fileProxies, err := networking.LoadProxiesFromFile(cfg.ProxyFile)
				if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading proxy file: %v\n", err)
			os.Exit(1)
		}
		proxyStrings = append(proxyStrings, fileProxies...)
	}

	if len(proxyStrings) > 0 {
		var parsedProxies []*url.URL
		for _, pStr := range proxyStrings {
			pURL, err := networking.ParseProxyURL(pStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not parse proxy '%s', skipping: %v\n", pStr, err)
				continue
			}
			parsedProxies = append(parsedProxies, pURL)
		}
		
		cfg.LoadedProxies = networking.CheckProxies(parsedProxies, cfg.Timeout, cfg.NoColor)

		if len(cfg.LoadedProxies) == 0 {
			fmt.Fprintf(os.Stderr, "Error: No live proxies found from the provided list. Aborting.\n")
			os.Exit(1)
		}
	}
	// --- End Proxy Setup ---

	// Add files from directory to targets
	if cfg.Directory != "" {
		err := filepath.Walk(cfg.Directory, func(path string, info os.FileInfo, err error) error {
			if err != nil { return err }
			// Extended file types for Harpy
			if !info.IsDir() {
				ext := strings.ToLower(filepath.Ext(info.Name()))
				if ext == ".js" || ext == ".ts" || ext == ".html" || ext == ".json" {
					cfg.Targets = append(cfg.Targets, path)
				}
			}
			return nil
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking directory %s: %v\n", cfg.Directory, err)
			os.Exit(1)
		}
	}

		if len(cfg.Targets) == 0 {
		fmt.Fprintln(os.Stderr, "No targets to scan. Provide targets via -u, -f, -d, or stdin.")
		os.Exit(0)
	}

	// Setup components
	terminalController := output.NewTerminalController()
	logger := utils.NewLogger(terminalController, cfg.Verbose)
	
	logger.Infof("Harpy starting...")
	
	progBar := output.NewProgressBar(terminalController)
	
	client, err := networking.NewClient(
		logger,
		cfg.Timeout,
		cfg.InsecureSkipVerify,
		cfg.Headers,
		cfg.LoadedProxies,
	)
	if err != nil {
		logger.Fatalf("Failed to create networking client: %v", err)
	}

	analyzer := core.NewAnalyzer(cfg, logger) // Changed from processor to analyzer
	domainManager := networking.NewDomainManager(cfg, logger)
	reporter := report.NewReporter(cfg, logger)
	scheduler := core.NewScheduler(cfg, client, analyzer, domainManager, logger, reporter, progBar) // Updated parameters

	analyzer.SetScheduler(scheduler) // Changed from processor.SetScheduler
	logger.SetProgressBar(progBar)

	// Log initial statistics
	logger.Infof("Loaded %d targets to scan.", len(cfg.Targets))
	if len(cfg.Targets) > 0 && strings.HasPrefix(cfg.Targets[0], "http") {
		uniqueDomains := make(map[string]struct{})
		for _, target := range cfg.Targets {
			if u, err := url.Parse(target); err == nil {
				uniqueDomains[u.Hostname()] = struct{}{}
			}
		}
		logger.Infof("Scanning across %d unique domains.", len(uniqueDomains))
	}

	logInitialSettings(logger, cfg)

	// Start the scan
	startTime := time.Now()
	logger.Infof("Starting scan with %d workers.", cfg.Concurrency)

	// Os workers devem ser iniciados ANTES de enfileirar os jobs
	// para evitar deadlocks se a lista de alvos for maior que o buffer do canal.
	scheduler.StartScan()

	scheduler.AddInitialTargets(cfg.Targets)
	progBar.Start(len(cfg.Targets))
	logger.SetProgBarActive(true)

	scheduler.Wait()

	progBar.Stop()
	logger.SetProgBarActive(false)

	// Print results
	reporter.Print()

	logger.Infof("Scan finished in %s.", time.Since(startTime).Round(time.Second))
	if reporter.GetFindingsCount() == 0 {
		logger.Infof("No findings extracted.")
	}
}

func logInitialSettings(logger *utils.Logger, cfg *config.Config) {
	settings := []string{
		fmt.Sprintf("Rate Limit: Auto (Max %d/s)", cfg.MaxRateLimit),
	}
	if cfg.NoLimit {
		settings = append(settings, "File Size Limit: Disabled")
	} else {
		settings = append(settings, fmt.Sprintf("File Size Limit: %d KB", cfg.MaxFileSize))
	}
	// Harpy always uses hybrid analysis with intelligent fallbacks
	settings = append(settings, "Analysis: Hybrid (Regex + AST with Smart Fallbacks)")
	if cfg.InsecureSkipVerify {
		settings = append(settings, "TLS Verification: Disabled")
	}
	if len(cfg.Headers) > 0 {
		settings = append(settings, fmt.Sprintf("Custom Headers: %d", len(cfg.Headers)))
	}
	if len(cfg.LoadedProxies) > 0 {
		settings = append(settings, fmt.Sprintf("Using %d Live Proxies", len(cfg.LoadedProxies)))
	} else if cfg.ProxyFile != "" || cfg.Proxy != "" {
		settings = append(settings, "Proxy: None (check failed)")
	}

	logger.Infof("Settings: %s", strings.Join(settings, ", "))
} 