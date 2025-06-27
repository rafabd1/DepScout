package report

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"DepScout/internal/config"
	"DepScout/internal/utils"
)

// Status constants for Findings
const (
	StatusConfirmed = "Confirmed"
	StatusPotential = "Potential"
	StatusReflected = "Reflected"
)

// Finding represents a single unclaimed package found.
type Finding struct {
	UnclaimedPackage string `json:"unclaimed_package"`
	FoundInSourceURL string `json:"source_url"`
}

// Reporter manages the collection and output of findings.
type Reporter struct {
	config     *config.Config
	logger     *utils.Logger
	findings   []Finding
	mu         sync.Mutex
	outputFile *os.File
}

// NewReporter creates a new Reporter.
func NewReporter(cfg *config.Config, logger *utils.Logger) *Reporter {
	var outputFile *os.File
	var err error

	if cfg.OutputFile != "" {
		outputFile, err = os.Create(cfg.OutputFile)
		if err != nil {
			logger.Fatalf("Failed to create output file %s: %v", cfg.OutputFile, err)
		}
	}

	return &Reporter{
		config:     cfg,
		logger:     logger,
		findings:   []Finding{},
		outputFile: outputFile,
	}
}

// AddFinding records a new finding.
func (r *Reporter) AddFinding(finding Finding) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.findings = append(r.findings, finding)

	if r.config.JsonOutput {
		jsonLine, err := json.Marshal(finding)
		if err == nil {
			if r.outputFile != nil {
				r.outputFile.Write(append(jsonLine, '\n'))
			} else {
				// If no output file, print JSON to stdout
				fmt.Println(string(jsonLine))
			}
		}
	} else if r.outputFile != nil {
		// Plain text output to file
		line := fmt.Sprintf("Package: %s, Source: %s\n", finding.UnclaimedPackage, finding.FoundInSourceURL)
		r.outputFile.WriteString(line)
	}
}

// GetFindingsCount returns the number of findings.
func (r *Reporter) GetFindingsCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.findings)
}

// Print displays the final report to the console.
func (r *Reporter) Print() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// If JSON output is to stdout, it has already been printed.
	if r.config.JsonOutput && r.outputFile == nil {
		return
	}

	if len(r.findings) > 0 {
		r.logger.Infof("Found %d unclaimed dependencies:", len(r.findings))
		for _, f := range r.findings {
			r.logger.Successf("  [+] Package: %s", f.UnclaimedPackage)
			if r.config.Verbose {
				r.logger.Infof("      Source:  %s", f.FoundInSourceURL)
			}
		}
	}
}

// Close safely closes the output file if it was opened.
func (r *Reporter) Close() {
	if r.outputFile != nil {
		r.outputFile.Close()
	}
} 