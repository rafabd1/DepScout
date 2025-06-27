package report

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
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

// Finding represents an unclaimed dependency discovery.
type Finding struct {
	UnclaimedPackage string `json:"unclaimed_package"`
	FoundInSourceURL string `json:"source_url"`
}

// Reporter handles the generation and output of scan results.
type Reporter struct {
	config   *config.Config
	logger   utils.Logger
	findings []Finding
	mu       sync.Mutex
}

// NewReporter creates a new Reporter.
func NewReporter(cfg *config.Config, logger utils.Logger) *Reporter {
	return &Reporter{
		config:   cfg,
		logger:   logger,
		findings: make([]Finding, 0),
	}
}

// AddFinding adds a new finding to the reporter in a thread-safe manner.
func (r *Reporter) AddFinding(finding Finding) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.findings = append(r.findings, finding)
}

// GetFindings returns a copy of all findings.
func (r *Reporter) GetFindings() []Finding {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Return a copy to prevent race conditions on the slice itself
	findingsCopy := make([]Finding, len(r.findings))
	copy(findingsCopy, r.findings)
	return findingsCopy
}

// PrintReport outputs all recorded findings to stdout.
func (r *Reporter) PrintReport() {
	findings := r.GetFindings() // Use thread-safe getter
	if len(findings) == 0 {
		if !r.config.Silent {
			r.logger.Infof("No unclaimed dependencies found.")
		}
		return
	}

	if strings.ToLower(r.config.OutputFormat) == "json" {
		output, err := json.MarshalIndent(findings, "", "  ")
		if err != nil {
			r.logger.Errorf("Failed to generate JSON report: %v", err)
			return
		}
		fmt.Println(string(output))
	} else {
		r.logger.Infof("Found %d unclaimed dependencies:", len(findings))
		for _, finding := range findings {
			// Simple text format
			fmt.Printf("  [+] Package: %s\n      Source:  %s\n", finding.UnclaimedPackage, finding.FoundInSourceURL)
		}
	}
}

// WriteReportToFile saves the findings to the specified output file.
func (r *Reporter) WriteReportToFile() error {
	findings := r.GetFindings() // Use thread-safe getter
	if len(findings) == 0 {
		return nil // Nothing to write
	}

	var output []byte
	var err error

	if strings.ToLower(r.config.OutputFormat) == "json" {
		output, err = json.MarshalIndent(findings, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to encode findings to JSON: %w", err)
		}
	} else {
		var sb strings.Builder
		for _, finding := range findings {
			sb.WriteString(fmt.Sprintf("Package: %s, Source: %s\n", finding.UnclaimedPackage, finding.FoundInSourceURL))
		}
		output = []byte(sb.String())
	}

	err = os.WriteFile(r.config.Output, output, 0644)
	if err != nil {
		return fmt.Errorf("failed to write report to file '%s': %w", r.config.Output, err)
	}

	r.logger.Infof("Report successfully saved to %s", r.config.Output)
	return nil
} 