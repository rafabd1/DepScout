package main

import (
	"fmt"
	"os"
	"time"

	"DepScout/internal/config"
	"DepScout/internal/core"
	"DepScout/internal/input"
	"DepScout/internal/networking"
	"DepScout/internal/output"
	"DepScout/internal/report"
	"DepScout/internal/utils"

	"github.com/spf13/cobra"
)

var (
	cfg    *config.Config
	logger utils.Logger
	version = "dev"
)

func main() {
	cmd := NewRootCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func NewRootCommand() *cobra.Command {
	cfg = config.NewDefaultConfig()

	rootCmd := &cobra.Command{
		Use:   "depscout",
		Short: "A tool to find potential dependency confusion vulnerabilities.",
		Long: `DepScout - Dependency Confusion Scanner

A high-performance security tool to detect potential Dependency Confusion vulnerabilities
by scanning JavaScript files for privately-named packages that are unclaimed on public registries.`,
		Example: `  # Scan a single remote JavaScript file:
  depscout -u https://example.com/assets/app.js

  # Scan a list of URLs from a file with 50 concurrent workers and JSON output:
  depscout -f /path/to/urls.txt -c 50 -o results.json --json
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
			utils.SetTerminalController(output.GetTerminalController())
			
			logLevel := utils.LevelInfo
			if cfg.Verbose {
				logLevel = utils.LevelDebug
			}
			logger = utils.NewDefaultLogger(logLevel, cfg.NoColor, cfg.Silent)

			if !cfg.Silent {
				fmt.Println("... DepScout banner ...")
			}
			
			var targets []string
			if cfg.URL != "" {
				targets = append(targets, cfg.URL)
			}
			if cfg.File != "" {
				fileTargets, err := input.LoadLinesFromFile(cfg.File)
				if err != nil {
					logger.Fatalf("Error reading targets file: %v", err)
				}
				targets = append(targets, fileTargets...)
			}
			
			if len(targets) == 0 {
				return fmt.Errorf("no targets specified, use -u or -f")
			}
			cfg.Targets = targets

			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("config validation error: %v", err)
			}
			
			if err := cfg.ParseAndValidateProxies(); err != nil {
				return fmt.Errorf("proxy config error: %v", err)
			}

			logger.Infof("Starting scan with %d workers.", cfg.Concurrency)

			client, _ := networking.NewClient(cfg, logger)
			processor := core.NewProcessor(cfg, logger)
			domainManager := networking.NewDomainManager(cfg, logger)
			reporter := report.NewReporter(cfg, logger)
			scheduler := core.NewScheduler(cfg, client, processor, domainManager, logger, reporter)
			
			processor.SetScheduler(scheduler)

			scheduler.AddInitialTargets(cfg.Targets)
			scheduler.StartScan()

			if !cfg.Silent {
				findings := reporter.GetFindings()
				if len(findings) > 0 {
					reporter.PrintReport()
				} else {
					logger.Infof("Scan finished. No unclaimed dependencies found.")
				}
			}
			
			if cfg.Output != "" {
				err := reporter.WriteReportToFile()
				if err != nil {
					logger.Fatalf("Failed to write findings to output file: %v", err)
				}
			}

			return nil
		},
	}

	// Flags
	rootCmd.Flags().StringVarP(&cfg.URL, "url", "u", "", "Single URL to scan")
	rootCmd.Flags().StringVarP(&cfg.File, "file", "f", "", "File with URLs")
	rootCmd.Flags().StringSliceVarP(&cfg.CustomHeaders, "header", "H", []string{}, "Custom headers")
	rootCmd.Flags().IntVarP(&cfg.Concurrency, "concurrency", "c", 25, "Concurrency level")
	rootCmd.Flags().DurationVarP(&cfg.RequestTimeout, "timeout", "t", 10*time.Second, "Request timeout")
	rootCmd.Flags().StringVarP(&cfg.Output, "output", "o", "", "Output file")
	rootCmd.Flags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Verbose output")
	rootCmd.Flags().BoolVarP(&cfg.Silent, "silent", "s", false, "Silent mode")
	rootCmd.Flags().StringVarP(&cfg.OutputFormat, "format", "", "text", "Output format (text or json)")
	rootCmd.Flags().StringVar(&cfg.Proxy, "proxy", "", "Single proxy URL")
	rootCmd.Flags().StringVar(&cfg.ProxyFile, "proxy-file", "", "File with list of proxies")
	rootCmd.Flags().IntVarP(&cfg.MaxRetries, "max-retries", "r", 2, "Max retries for failed requests")
	rootCmd.Flags().BoolVar(&cfg.NoColor, "no-color", false, "Disable color output")
	rootCmd.Flags().BoolVarP(&cfg.InsecureSkipVerify, "insecure", "k", false, "Skip TLS certificate verification")
	rootCmd.Flags().Int64Var(&cfg.MaxFileSize, "max-file-size", 1048576, "Max file size to process in bytes (default 1MB)")

	return rootCmd
} 