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
	Source     string              `json:"source"`
	Domains    []string           `json:"domains"`
	Endpoints  []EndpointFinding  `json:"endpoints"`
	Parameters []ParameterFinding `json:"parameters"`
	Headers    []HeaderFinding    `json:"headers"`
}

type EndpointFinding struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Context string `json:"context"`
}

type ParameterFinding struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Context  string `json:"context"`
}

type HeaderFinding struct {
	Name    string `json:"name"`
	Context string `json:"context"`
}

// HarpyOutput represents the complete JSON output structure
type HarpyOutput struct {
	Metadata HarpyMetadata   `json:"metadata"`
	Results  HarpyResults    `json:"results"`
	Summary  HarpySummary    `json:"summary"`
}

type HarpyMetadata struct {
	Tool        string `json:"tool"`
	Version     string `json:"version"`
	Timestamp   string `json:"timestamp"`
	TotalSources int   `json:"total_sources"`
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
	config        *config.Config
	logger        *utils.Logger
	harpyFindings []HarpyFinding // Harpy-specific findings
	mu            sync.Mutex
	outputFile    *os.File
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
		config:        cfg,
		logger:        logger,
		harpyFindings: []HarpyFinding{}, // Initialize Harpy findings
		outputFile:    outputFile,
	}
}

// AddHarpyFinding records a new Harpy finding.
func (r *Reporter) AddHarpyFinding(finding HarpyFinding) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.harpyFindings = append(r.harpyFindings, finding)
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

	// Handle file output (JSON or text based on config)
	if r.outputFile != nil {
		if r.config.JsonOutput {
			// Create structured JSON output
			summary := r.calculateSummary()
			output := HarpyOutput{
				Metadata: HarpyMetadata{
					Tool:         "Harpy",
					Version:      "1.0.0",
					Timestamp:    time.Now().Format(time.RFC3339),
					TotalSources: len(r.harpyFindings),
				},
				Results: HarpyResults{
					Findings: r.harpyFindings,
				},
				Summary: summary,
			}
			
			encoder := json.NewEncoder(r.outputFile)
			encoder.SetIndent("", "  ")
			err := encoder.Encode(output)
			if err != nil {
				r.logger.Errorf("Failed to write JSON output to file: %v", err)
			}
		} else {
			r.writeTextOutput()
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
		r.outputFile.Close()
	}
}

// writeTextOutput writes findings to file in text format
func (r *Reporter) writeTextOutput() {
	totalFindings := len(r.harpyFindings)
	if totalFindings == 0 {
		return
	}
	
	r.outputFile.WriteString("=== Harpy Results ===\n")
	r.outputFile.WriteString(fmt.Sprintf("Total findings: %d\n\n", totalFindings))
	
	// Write Harpy findings
	if len(r.harpyFindings) > 0 {
		r.outputFile.WriteString(fmt.Sprintf("Extracted Data (%d sources):\n", len(r.harpyFindings)))
		for _, f := range r.harpyFindings {
			r.outputFile.WriteString(fmt.Sprintf("\nSource: %s\n", f.Source))
			
			if len(f.Domains) > 0 {
				r.outputFile.WriteString(fmt.Sprintf("  Domains (%d):\n", len(f.Domains)))
				for _, d := range f.Domains {
					r.outputFile.WriteString(fmt.Sprintf("    - %s\n", d))
				}
			}
			
			if len(f.Endpoints) > 0 {
				r.outputFile.WriteString(fmt.Sprintf("  Endpoints (%d):\n", len(f.Endpoints)))
				for _, e := range f.Endpoints {
					r.outputFile.WriteString(fmt.Sprintf("    - %s %s\n", e.Method, e.Path))
				}
			}
			
			if len(f.Parameters) > 0 {
				r.outputFile.WriteString(fmt.Sprintf("  Parameters (%d):\n", len(f.Parameters)))
				for _, p := range f.Parameters {
					r.outputFile.WriteString(fmt.Sprintf("    - %s (%s)\n", p.Name, p.Type))
				}
			}
			
			if len(f.Headers) > 0 {
				r.outputFile.WriteString(fmt.Sprintf("  Headers (%d):\n", len(f.Headers)))
				for _, h := range f.Headers {
					r.outputFile.WriteString(fmt.Sprintf("    - %s\n", h.Name))
				}
			}
		}
	}
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