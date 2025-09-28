package report

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/rafabd1/DepScout/internal/config"
	"github.com/rafabd1/DepScout/internal/utils"
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
}

// GetFindingsCount returns the number of findings.
func (r *Reporter) GetFindingsCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.findings)
}

// Print displays the final report.
func (r *Reporter) Print() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.config.Silent && len(r.findings) == 0 {
		return
	}

	// Handle file output (JSON or text based on config)
	if r.outputFile != nil {
		if r.config.JsonOutput {
			// Write JSON to file
			encoder := json.NewEncoder(r.outputFile)
			encoder.SetIndent("", "  ")
			err := encoder.Encode(r.findings)
			if err != nil {
				r.logger.Errorf("Failed to write JSON output to file: %v", err)
			}
		} else {
			// Write text format to file
			if len(r.findings) > 0 {
				r.outputFile.WriteString(fmt.Sprintf("Found %d unclaimed dependencies:\n", len(r.findings)))
				for _, f := range r.findings {
					r.outputFile.WriteString(fmt.Sprintf("  [+] Package: %s\n", f.UnclaimedPackage))
					if r.config.Verbose {
						r.outputFile.WriteString(fmt.Sprintf("      Source:  %s\n", f.FoundInSourceURL))
					}
				}
			}
		}
	}

	// Always display human-readable format in terminal (unless silent)
	if !r.config.Silent {
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
}

// Close safely closes the output file if it was opened.
func (r *Reporter) Close() {
	if r.outputFile != nil {
		r.outputFile.Close()
	}
} 