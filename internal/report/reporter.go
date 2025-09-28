package report

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rafabd1/Harpy/internal/config"
	"github.com/rafabd1/Harpy/internal/utils"
)

// Status constants for Findings
const (
	StatusConfirmed = "Confirmed"
	StatusPotential = "Potential"
	StatusReflected = "Reflected"
)

// HarpyFinding represents findings extracted by Harpy
type HarpyFinding struct {
	Source     string             `json:"source"`
	Domains    []string           `json:"domains"`
	Endpoints  []EndpointFinding  `json:"endpoints"`
	Parameters []ParameterFinding `json:"parameters"`
	Headers    []HeaderFinding    `json:"headers"`
}

type EndpointFinding struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Context string `json:"context"`
}

type ParameterFinding struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Context string `json:"context"`
}

type HeaderFinding struct {
	Name    string `json:"name"`
	Context string `json:"context"`
}

// HarpyOutput represents the complete JSON output structure
type HarpyOutput struct {
	Metadata HarpyMetadata `json:"metadata"`
	Results  HarpyResults  `json:"results"`
	Summary  HarpySummary  `json:"summary"`
}

type HarpyMetadata struct {
	Tool         string `json:"tool"`
	Version      string `json:"version"`
	Timestamp    string `json:"timestamp"`
	TotalSources int    `json:"total_sources"`
}

type HarpyResults struct {
	Findings []HarpyFinding `json:"findings"`
}

type HarpySummary struct {
	TotalDomains    int `json:"total_domains"`
	TotalEndpoints  int `json:"total_endpoints"`
	TotalParameters int `json:"total_parameters"`
	TotalHeaders    int `json:"total_headers"`
}

// Reporter manages the collection and output of findings.
type Reporter struct {
	config           *config.Config
	logger           *utils.Logger
	harpyFindings    []HarpyFinding // Harpy-specific findings
	mu               sync.Mutex
	outputFile       *os.File
	isJsonOutput     bool
	hasWrittenHeader bool
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

	reporter := &Reporter{
		config:        cfg,
		logger:        logger,
		harpyFindings: []HarpyFinding{}, // Initialize Harpy findings
		outputFile:    outputFile,
		isJsonOutput:  cfg.JsonOutput,
	}

	// Write JSON header if JSON output is enabled
	if outputFile != nil && cfg.JsonOutput {
		reporter.writeJsonHeader()
	}

	return reporter
}

// AddHarpyFinding records a new Harpy finding.
func (r *Reporter) AddHarpyFinding(finding HarpyFinding) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if finding has any useful data
	if len(finding.Domains) == 0 && len(finding.Endpoints) == 0 &&
		len(finding.Parameters) == 0 && len(finding.Headers) == 0 {

		return // Skip empty findings
	}

	r.harpyFindings = append(r.harpyFindings, finding)

	// Write finding to file immediately if file output is enabled
	if r.outputFile != nil {

		if r.isJsonOutput {
			r.writeJsonFinding(finding)
		} else {
			r.writeTextFinding(finding)
		}
		r.outputFile.Sync()

	}
}

// GetFindingsCount returns the total number of findings.
func (r *Reporter) GetFindingsCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.harpyFindings)
}

// GetHarpyFindingsCount returns the number of Harpy findings.
func (r *Reporter) GetHarpyFindingsCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.harpyFindings)
}

// Print displays the final report.
func (r *Reporter) Print() {
	r.mu.Lock()
	defer r.mu.Unlock()

	totalFindings := len(r.harpyFindings)

	if r.config.Silent && totalFindings == 0 {
		return
	}

	// File output is handled in real-time now, so we only need to finalize
	if r.outputFile != nil {
		if r.config.JsonOutput {
			// JSON footer is written automatically when closing
		} else {
			// For text output, add final summary
			r.outputFile.WriteString("\n=== Final Summary ===\n")
			r.outputFile.WriteString(fmt.Sprintf("Total sources processed: %d\n", len(r.harpyFindings)))
			summary := r.calculateSummary()
			if summary.TotalDomains > 0 {
				r.outputFile.WriteString(fmt.Sprintf("Total domains found: %d\n", summary.TotalDomains))
			}
			if summary.TotalEndpoints > 0 {
				r.outputFile.WriteString(fmt.Sprintf("Total endpoints found: %d\n", summary.TotalEndpoints))
			}
			if summary.TotalParameters > 0 {
				r.outputFile.WriteString(fmt.Sprintf("Total parameters found: %d\n", summary.TotalParameters))
			}
			if summary.TotalHeaders > 0 {
				r.outputFile.WriteString(fmt.Sprintf("Total headers found: %d\n", summary.TotalHeaders))
			}
			r.outputFile.Sync() // Force final write
		}
	}

	// Always display human-readable format in terminal (unless silent)
	if !r.config.Silent {
		r.printToTerminal()
	}
}

// Close safely closes the output file if it was opened.
func (r *Reporter) Close() {
	if r.outputFile != nil {
		if r.isJsonOutput {
			r.writeJsonFooter()
		}
		r.outputFile.Close()
	}
}

// writeTextOutput is deprecated - findings are now written in real-time
// This function is kept for compatibility but does nothing
func (r *Reporter) writeTextOutput() {
	// Real-time writing is handled in writeTextFinding
	// This method is now a no-op
}

// printToTerminal prints clean formatted output to terminal
func (r *Reporter) printToTerminal() {
	totalFindings := len(r.harpyFindings)
	if totalFindings == 0 {
		r.logger.Infof("No findings extracted.")
		return
	}

	// Summary statistics
	totalDomains := 0
	totalEndpoints := 0
	totalParameters := 0
	totalHeaders := 0

	for _, f := range r.harpyFindings {
		totalDomains += len(f.Domains)
		totalEndpoints += len(f.Endpoints)
		totalParameters += len(f.Parameters)
		totalHeaders += len(f.Headers)
	}

	r.logger.Infof("=== Extraction Results ===")
	r.logger.Successf("Sources processed: %d", len(r.harpyFindings))
	if totalDomains > 0 {
		r.logger.Successf("Domains found: %d", totalDomains)
	}
	if totalEndpoints > 0 {
		r.logger.Successf("Endpoints found: %d", totalEndpoints)
	}
	if totalParameters > 0 {
		r.logger.Successf("Parameters found: %d", totalParameters)
	}
	if totalHeaders > 0 {
		r.logger.Successf("Headers found: %d", totalHeaders)
	}

	// Detailed output in verbose mode
	if r.config.Verbose && len(r.harpyFindings) > 0 {
		r.logger.Infof("\n=== Detailed Findings ===")
		for _, f := range r.harpyFindings {
			r.logger.Infof("Source: %s", f.Source)

			if len(f.Domains) > 0 {
				for _, d := range f.Domains {
					r.logger.Successf("  [DOMAIN] %s", d)
				}
			}

			if len(f.Endpoints) > 0 {
				for _, e := range f.Endpoints {
					r.logger.Successf("  [ENDPOINT] %s %s", e.Method, e.Path)
				}
			}

			if len(f.Parameters) > 0 {
				for _, p := range f.Parameters {
					r.logger.Successf("  [PARAM] %s (%s)", p.Name, p.Type)
				}
			}

			if len(f.Headers) > 0 {
				for _, h := range f.Headers {
					r.logger.Successf("  [HEADER] %s", h.Name)
				}
			}
		}
	}
}

// calculateSummary calculates summary statistics from findings
func (r *Reporter) calculateSummary() HarpySummary {
	summary := HarpySummary{}

	for _, f := range r.harpyFindings {
		summary.TotalDomains += len(f.Domains)
		summary.TotalEndpoints += len(f.Endpoints)
		summary.TotalParameters += len(f.Parameters)
		summary.TotalHeaders += len(f.Headers)
	}

	return summary
}

// writeJsonHeader writes the beginning of JSON output structure
func (r *Reporter) writeJsonHeader() {
	if r.outputFile == nil {
		return
	}

	metadata := HarpyMetadata{
		Tool:      "Harpy",
		Version:   "1.0.0",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	r.outputFile.WriteString("{\n")
	r.outputFile.WriteString("  \"metadata\": {\n")
	r.outputFile.WriteString(fmt.Sprintf("    \"tool\": \"%s\",\n", metadata.Tool))
	r.outputFile.WriteString(fmt.Sprintf("    \"version\": \"%s\",\n", metadata.Version))
	r.outputFile.WriteString(fmt.Sprintf("    \"timestamp\": \"%s\"\n", metadata.Timestamp))
	r.outputFile.WriteString("  },\n")
	r.outputFile.WriteString("  \"results\": {\n    \"findings\": [\n")
	r.outputFile.Sync() // Force write to disk
	r.hasWrittenHeader = true
}

// writeJsonFinding writes a single finding to JSON output
func (r *Reporter) writeJsonFinding(finding HarpyFinding) {
	if r.outputFile == nil || !r.hasWrittenHeader {
		return
	}

	// Add comma separator for subsequent findings
	if len(r.harpyFindings) > 1 {
		r.outputFile.WriteString(",\n")
	}

	jsonBytes, _ := json.MarshalIndent(finding, "      ", "  ")
	r.outputFile.WriteString(string(jsonBytes))
}

// writeJsonFooter writes the end of JSON output structure
func (r *Reporter) writeJsonFooter() {
	if r.outputFile == nil || !r.hasWrittenHeader {
		return
	}

	summary := r.calculateSummary()

	r.outputFile.WriteString("\n    ]\n  },\n")
	r.outputFile.WriteString("  \"summary\": {\n")
	r.outputFile.WriteString(fmt.Sprintf("    \"total_domains\": %d,\n", summary.TotalDomains))
	r.outputFile.WriteString(fmt.Sprintf("    \"total_endpoints\": %d,\n", summary.TotalEndpoints))
	r.outputFile.WriteString(fmt.Sprintf("    \"total_parameters\": %d,\n", summary.TotalParameters))
	r.outputFile.WriteString(fmt.Sprintf("    \"total_headers\": %d\n", summary.TotalHeaders))
	r.outputFile.WriteString("  }\n}\n")
}

// writeTextFinding writes a single finding to text output
func (r *Reporter) writeTextFinding(finding HarpyFinding) {
	if r.outputFile == nil {
		return
	}

	// Write header only once
	if len(r.harpyFindings) == 1 {
		r.outputFile.WriteString("=== Harpy Results (Real-time) ===\n\n")
	}

	r.outputFile.WriteString(fmt.Sprintf("Source: %s\n", finding.Source))

	if len(finding.Domains) > 0 {
		r.outputFile.WriteString(fmt.Sprintf("  Domains (%d):\n", len(finding.Domains)))
		for _, d := range finding.Domains {
			r.outputFile.WriteString(fmt.Sprintf("    - %s\n", d))
		}
	}

	if len(finding.Endpoints) > 0 {
		r.outputFile.WriteString(fmt.Sprintf("  Endpoints (%d):\n", len(finding.Endpoints)))
		for _, e := range finding.Endpoints {
			r.outputFile.WriteString(fmt.Sprintf("    - %s %s\n", e.Method, e.Path))
		}
	}

	if len(finding.Parameters) > 0 {
		r.outputFile.WriteString(fmt.Sprintf("  Parameters (%d):\n", len(finding.Parameters)))
		for _, p := range finding.Parameters {
			r.outputFile.WriteString(fmt.Sprintf("    - %s (%s)\n", p.Name, p.Type))
		}
	}

	if len(finding.Headers) > 0 {
		r.outputFile.WriteString(fmt.Sprintf("  Headers (%d):\n", len(finding.Headers)))
		for _, h := range finding.Headers {
			r.outputFile.WriteString(fmt.Sprintf("    - %s\n", h.Name))
		}
	}

	r.outputFile.WriteString("\n")
}
