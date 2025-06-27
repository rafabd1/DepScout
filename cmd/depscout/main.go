package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"DepScout/internal/config"
	"DepScout/internal/core"
	"DepScout/internal/networking"
	"DepScout/internal/output"
	"DepScout/internal/report"
	"DepScout/internal/utils"
)

func main() {
	cfg := config.NewConfig()
	if err := cfg.Parse(); err != nil {
		if err != flag.ErrHelp {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Fprintln(os.Stderr, "Use -h or --help for usage.")
		}
		os.Exit(1)
	}

	// Add files from directory to targets
	if cfg.Directory != "" {
		err := filepath.Walk(cfg.Directory, func(path string, info os.FileInfo, err error) error {
			if err != nil { return err }
			if !info.IsDir() && (strings.HasSuffix(strings.ToLower(info.Name()), ".js") || strings.HasSuffix(strings.ToLower(info.Name()), ".ts")) {
				cfg.Targets = append(cfg.Targets, path)
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
	
	logger.Infof("DepScout starting...")
	
	progBar := output.NewProgressBar(terminalController)
	
	client, err := networking.NewClient(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to create networking client: %v", err)
	}

	processor := core.NewProcessor(cfg, logger)
	domainManager := networking.NewDomainManager(cfg, logger)
	reporter := report.NewReporter(cfg, logger)
	scheduler := core.NewScheduler(cfg, client, processor, domainManager, logger, reporter, progBar)

	processor.SetScheduler(scheduler)

	// Start the scan
	startTime := time.Now()
	logger.Infof("Starting scan with %d workers.", cfg.Concurrency)

	scheduler.AddInitialTargets(cfg.Targets)
	progBar.Start(len(cfg.Targets))
	
	scheduler.StartScan()
	scheduler.Wait()

	progBar.Stop()
	
	// Print results
	reporter.Print()

	logger.Infof("Scan finished in %s.", time.Since(startTime).Round(time.Second))
	if reporter.GetFindingsCount() == 0 {
		logger.Infof("No unclaimed dependencies found.")
	}
} 