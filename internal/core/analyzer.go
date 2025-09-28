package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/rafabd1/Harpy/internal/config"
	"github.com/rafabd1/Harpy/internal/report"
	"github.com/rafabd1/Harpy/internal/utils"
)

/**
 * @description Hybrid Analysis Engine - Core component that combines regex and AST analysis
 * for intelligent endpoint and parameter extraction from web applications
 */
type Analyzer struct {
	config        *config.Config
	logger        *utils.Logger
	regexEngine   *RegexEngine
	astEngine     *ASTEngine
	patternDB     *PatternDatabase
	contextMapper *ContextMapper
	scheduler     *Scheduler
	processed     sync.Map // Tracks processed content to avoid duplicates
}

/**
 * @description Creates a new Analyzer instance with all extraction engines initialized
 * @param cfg Configuration for analysis behavior
 * @param logger Logger instance for debug/info output
 * @returns Configured Analyzer ready for processing
 */
func NewAnalyzer(cfg *config.Config, logger *utils.Logger) *Analyzer {
	regexEngine := NewRegexEngine()
	astEngine := NewASTEngine()
	patternDB := NewPatternDatabase()
	contextMapper := NewContextMapper(500) // 500 characters proximity threshold

	return &Analyzer{
		config:        cfg,
		logger:        logger,
		regexEngine:   regexEngine,
		astEngine:     astEngine,
		patternDB:     patternDB,
		contextMapper: contextMapper,
	}
}

/**
 * @description Sets the scheduler reference for adding new jobs during analysis
 * @param s Scheduler instance for job management
 */
func (a *Analyzer) SetScheduler(s *Scheduler) {
	a.scheduler = s
}

/**
 * @description Main processing function that analyzes content using hybrid approach
 * @param sourceURL Source URL or file path of the content
 * @param content Raw content bytes to analyze
 * @returns Error if processing fails completely
 */
func (a *Analyzer) ProcessContent(sourceURL string, content []byte) error {
	// Skip empty or very small content
	if len(content) < 10 {
		return nil
	}

	// Check if already processed to avoid duplicates
	contentHash := a.generateContentHash(content)
	if _, exists := a.processed.LoadOrStore(contentHash, true); exists {
		return nil // Silently skip duplicates
	}

	// Determine file type for appropriate processing
	fileType := a.detectFileType(sourceURL, content)

	// Multi-pass analysis approach
	finding, err := a.processWithFallback(content, sourceURL, fileType)
	if err != nil {
		a.logger.Errorf("Failed to process %s: %v", sourceURL, err)
		return err
	}

	// Skip if no useful data was extracted
	if finding == nil || a.isEmpty(finding) {
		a.logger.Debugf("No useful findings extracted from %s", sourceURL)
		return nil // Silently skip empty results
	}

	// Add finding to scheduler for reporting
	if a.scheduler != nil {
		reportJob := Job{
			Input:   sourceURL,
			Type:    ReportFinding,
			Finding: finding,
		}
		if !a.scheduler.AddJobAsync(reportJob) {
			a.logger.Debugf("Could not queue reporting job for %s, adding directly to reporter", sourceURL)
			// Fallback: add directly to reporter if scheduler is shutting down
			a.addFindingDirectly(finding)
		}
	} else {
		a.logger.Warnf("Scheduler is nil, adding finding directly to reporter for %s", sourceURL)
		a.addFindingDirectly(finding)
	}

	return nil
}

// addFindingDirectly adds finding directly to reporter when scheduler is unavailable
func (a *Analyzer) addFindingDirectly(finding *Finding) {
	if a.scheduler == nil || a.scheduler.reporter == nil {
		return
	}

	// Convert core.Finding to report.HarpyFinding
	harpyFinding := report.HarpyFinding{
		Source:     finding.Source.FilePath,
		Domains:    finding.Domains,
		Endpoints:  make([]report.EndpointFinding, len(finding.Endpoints)),
		Parameters: make([]report.ParameterFinding, len(finding.Parameters)),
		Headers:    make([]report.HeaderFinding, len(finding.Headers)),
	}

	// Convert endpoints
	for i, ep := range finding.Endpoints {
		harpyFinding.Endpoints[i] = report.EndpointFinding{
			Method:  ep.Method,
			Path:    ep.Path,
			Context: ep.Context,
		}
	}

	// Convert parameters
	for i, param := range finding.Parameters {
		harpyFinding.Parameters[i] = report.ParameterFinding{
			Name:    param.Name,
			Type:    param.Type.String(),
			Context: param.Context,
		}
	}

	// Convert headers
	for i, header := range finding.Headers {
		harpyFinding.Headers[i] = report.HeaderFinding{
			Name:    header.Name,
			Context: header.Context,
		}
	}

	a.scheduler.reporter.AddHarpyFinding(harpyFinding)
}

/**
 * @description Processes content with multiple fallback strategies for resilience
 * @param content Content bytes to process
 * @param sourceURL Source identifier for context
 * @param fileType Detected file type for optimization
 * @returns Extracted finding or nil if nothing found
 */
func (a *Analyzer) processWithFallback(content []byte, sourceURL string, fileType FileType) (*Finding, error) {
	var finding *Finding
	var lastError error

	// Strategy selection based on file type and config
	strategies := a.selectStrategies(fileType)

	for i, strategy := range strategies {
		result, err := strategy(content, sourceURL)
		if err == nil && result != nil && !a.isEmpty(result) {
			finding = result
			break
		} else {
			lastError = err
			if err != nil {
				a.logger.Debugf("Analysis strategy %d failed for %s: %v", i+1, sourceURL, err)
			}
		}
	}

	if finding == nil {
		return nil, lastError
	}

	// Post-process to ensure data quality
	return a.validateAndCleanup(finding), nil
}

/**
 * @description Selects appropriate extraction strategies based on file type and config
 * @param fileType Detected file type
 * @returns Array of extraction strategy functions
 */
func (a *Analyzer) selectStrategies(fileType FileType) []func([]byte, string) (*Finding, error) {
	// Check configuration to determine which strategies to use
	strategies := make([]func([]byte, string) (*Finding, error), 0)

	switch fileType {
	case JavaScript, TypeScript:
		// For JS/TS, use hybrid if both are enabled, otherwise use available method
		if a.config.EnableRegex && a.config.EnableAST {
			strategies = append(strategies, a.processWithHybridJS)
		}
		if a.config.EnableRegex {
			strategies = append(strategies, a.processWithRegexOnly)
		}
		// Always have minimal extraction as last resort
		strategies = append(strategies, a.processWithMinimalExtraction)

	case HTML:
		strategies = append(strategies, a.processWithHTMLAnalysis)
		if a.config.EnableRegex {
			strategies = append(strategies, a.processWithRegexOnly)
		}

	case JSON:
		strategies = append(strategies, a.processWithJSONAnalysis)
		if a.config.EnableRegex {
			strategies = append(strategies, a.processWithRegexOnly)
		}

	default:
		// For unknown file types, use comprehensive extraction approach
		if a.config.EnableRegex {
			strategies = append(strategies, a.processWithRegexOnly)
		}
		// Add HTML analysis in case it's a web page without proper extension
		strategies = append(strategies, a.processWithHTMLAnalysis)
		// Add JSON analysis in case it's an API response or config
		strategies = append(strategies, a.processWithJSONAnalysis)
		// Always have minimal extraction as last resort
		strategies = append(strategies, a.processWithMinimalExtraction)
	}

	// Ensure we always have at least one strategy
	if len(strategies) == 0 {
		strategies = append(strategies, a.processWithMinimalExtraction)
	}

	return strategies
}

/**
 * @description Primary strategy: Hybrid JavaScript/TypeScript processing with regex + AST
 * @param content Content to analyze
 * @param sourceURL Source identifier
 * @returns Finding with extracted data
 */
func (a *Analyzer) processWithHybridJS(content []byte, sourceURL string) (*Finding, error) {
	// Ensure both regex and AST are enabled for hybrid processing
	if !a.config.EnableRegex || !a.config.EnableAST {
		return nil, fmt.Errorf("hybrid processing requires both regex and AST to be enabled")
	}

	// Pass 1: Fast regex extraction
	regexFindings := a.regexEngine.ExtractAll(content, sourceURL)
	if len(regexFindings) == 0 {
		return nil, nil
	}

	// Pass 2: AST enhancement for better context
	enhancedFindings := a.astEngine.EnhanceFindings(content, regexFindings)

	// Pass 3: Context mapping and relationship building
	mappedFindings := a.contextMapper.MapContext(enhancedFindings, content)

	// Pass 4: Merge and normalize
	return a.mergeFindings(mappedFindings, sourceURL), nil
}

/**
 * @description Fallback strategy: Regex-only processing
 * @param content Content to analyze
 * @param sourceURL Source identifier
 * @returns Finding with basic regex extraction
 */
func (a *Analyzer) processWithRegexOnly(content []byte, sourceURL string) (*Finding, error) {
	findings := a.regexEngine.ExtractAll(content, sourceURL)
	if len(findings) == 0 {
		return nil, nil
	}

	return a.mergeFindings(findings, sourceURL), nil
}

/**
 * @description Specialized HTML processing strategy
 * @param content HTML content to analyze
 * @param sourceURL Source identifier
 * @returns Finding with HTML-specific extraction
 */
func (a *Analyzer) processWithHTMLAnalysis(content []byte, sourceURL string) (*Finding, error) {
	// Extract from script tags, form actions, AJAX calls, etc.
	findings := a.regexEngine.ExtractFromHTML(content, sourceURL)
	if len(findings) == 0 {
		return nil, nil
	}

	return a.mergeFindings(findings, sourceURL), nil
}

/**
 * @description Specialized JSON processing strategy
 * @param content JSON content to analyze
 * @param sourceURL Source identifier
 * @returns Finding with JSON-specific extraction
 */
func (a *Analyzer) processWithJSONAnalysis(content []byte, sourceURL string) (*Finding, error) {
	// Extract from JSON configurations, API definitions, etc.
	findings := a.regexEngine.ExtractFromJSON(content, sourceURL)
	if len(findings) == 0 {
		return nil, nil
	}

	return a.mergeFindings(findings, sourceURL), nil
}

/**
 * @description Last resort strategy: Minimal URL pattern extraction
 * @param content Content to analyze
 * @param sourceURL Source identifier
 * @returns Finding with minimal data
 */
func (a *Analyzer) processWithMinimalExtraction(content []byte, sourceURL string) (*Finding, error) {
	// Extract at least basic URL patterns
	urls := a.regexEngine.ExtractURLPatterns(content)
	if len(urls) == 0 {
		return nil, nil
	}

	finding := &Finding{
		Source: SourceInfo{
			FilePath: sourceURL,
			FileType: a.detectFileType(sourceURL, content).String(),
		},
	}

	// Convert URL patterns to basic endpoints
	for _, urlPattern := range urls {
		endpoint := Endpoint{
			Path:    urlPattern,
			Method:  "GET", // Default assumption
			Context: "minimal_extraction",
		}
		finding.Endpoints = append(finding.Endpoints, endpoint)
	}

	return finding, nil
}

/**
 * @description Validates and cleans up extracted findings
 * @param finding Raw finding to validate
 * @returns Cleaned finding or nil if invalid
 */
func (a *Analyzer) validateAndCleanup(finding *Finding) *Finding {
	if finding == nil {
		return nil
	}

	// Remove invalid endpoints
	var validEndpoints []Endpoint
	for _, ep := range finding.Endpoints {
		if a.isValidEndpoint(ep) {
			validEndpoints = append(validEndpoints, ep)
		}
	}
	finding.Endpoints = validEndpoints

	// Validate parameters
	finding.Parameters = a.filterValidParameters(finding.Parameters)

	// Validate headers
	finding.Headers = a.filterValidHeaders(finding.Headers)

	// Ensure minimum data quality
	if len(finding.Endpoints) == 0 && len(finding.Parameters) == 0 &&
		len(finding.Headers) == 0 && len(finding.Domains) == 0 {
		return nil
	}

	return finding
}

/**
 * @description Checks if an endpoint is valid and worth reporting
 * @param ep Endpoint to validate
 * @returns True if endpoint is valid
 */
func (a *Analyzer) isValidEndpoint(ep Endpoint) bool {
	// Remove empty or invalid paths
	if ep.Path == "" || len(ep.Path) < 2 {
		return false
	}

	// Skip common false positives
	falsePaths := []string{"/", "//", "http://", "https://", "data:", "javascript:"}
	for _, falseP := range falsePaths {
		if ep.Path == falseP {
			return false
		}
	}

	// Must look like a valid path
	return strings.HasPrefix(ep.Path, "/") || strings.Contains(ep.Path, "://")
}

/**
 * @description Filters valid parameters from extracted list
 * @param params Parameters to filter
 * @returns Filtered valid parameters
 */
func (a *Analyzer) filterValidParameters(params []Parameter) []Parameter {
	var valid []Parameter
	seen := make(map[string]bool)

	for _, param := range params {
		if param.Name == "" || len(param.Name) < 1 {
			continue
		}

		// Avoid duplicates
		key := param.Name + "_" + param.Type.String()
		if seen[key] {
			continue
		}
		seen[key] = true

		valid = append(valid, param)
	}

	return valid
}

/**
 * @description Filters valid headers from extracted list
 * @param headers Headers to filter
 * @returns Filtered valid headers
 */
func (a *Analyzer) filterValidHeaders(headers []Header) []Header {
	var valid []Header
	seen := make(map[string]bool)

	for _, header := range headers {
		if header.Name == "" || len(header.Name) < 2 {
			continue
		}

		// Avoid duplicates
		if seen[header.Name] {
			continue
		}
		seen[header.Name] = true

		valid = append(valid, header)
	}

	return valid
}

// Helper functions

func (a *Analyzer) detectFileType(sourceURL string, content []byte) FileType {
	url := strings.ToLower(sourceURL)

	// For content analysis, use first 1000 characters for better detection
	contentStr := strings.ToLower(string(content[:min(1000, len(content))]))

	// First check URL-based detection (most reliable)
	switch {
	case strings.Contains(url, ".js"):
		return JavaScript
	case strings.Contains(url, ".ts") || strings.Contains(url, ".tsx"):
		return TypeScript
	case strings.Contains(url, ".html") || strings.Contains(url, ".htm"):
		return HTML
	case strings.Contains(url, ".json"):
		return JSON
	}

	// Enhanced content-based detection for URLs without extensions
	switch {
	// JSON detection - more comprehensive patterns
	case strings.HasPrefix(strings.TrimSpace(contentStr), "{") &&
		strings.HasSuffix(strings.TrimSpace(contentStr), "}"):
		return JSON
	case strings.HasPrefix(strings.TrimSpace(contentStr), "[") &&
		strings.HasSuffix(strings.TrimSpace(contentStr), "]"):
		return JSON
	case strings.Contains(contentStr, `"api"`) || strings.Contains(contentStr, `"endpoint"`) ||
		strings.Contains(contentStr, `"swagger"`) || strings.Contains(contentStr, `"openapi"`):
		return JSON

	// HTML detection
	case strings.Contains(contentStr, "<html") || strings.Contains(contentStr, "<!doctype") ||
		strings.Contains(contentStr, "<head>") || strings.Contains(contentStr, "<body>"):
		return HTML

	// TypeScript detection - look for TS-specific patterns
	case strings.Contains(contentStr, "interface ") || strings.Contains(contentStr, "type ") ||
		strings.Contains(contentStr, "export interface") || strings.Contains(contentStr, "export type") ||
		strings.Contains(contentStr, ": string") || strings.Contains(contentStr, ": number") ||
		strings.Contains(contentStr, "import type"):
		return TypeScript

	// JavaScript detection - comprehensive patterns
	case strings.Contains(contentStr, "function ") || strings.Contains(contentStr, "var ") ||
		strings.Contains(contentStr, "let ") || strings.Contains(contentStr, "const ") ||
		strings.Contains(contentStr, "=>") || strings.Contains(contentStr, "export ") ||
		strings.Contains(contentStr, "import ") || strings.Contains(contentStr, "require(") ||
		strings.Contains(contentStr, "module.exports") || strings.Contains(contentStr, ".prototype") ||
		strings.Contains(contentStr, "console.log") || strings.Contains(contentStr, "document.") ||
		strings.Contains(contentStr, "window.") || strings.Contains(contentStr, "fetch(") ||
		strings.Contains(contentStr, "axios.") || strings.Contains(contentStr, "$."):
		return JavaScript

	// API-like content detection - likely to contain endpoints
	case strings.Contains(contentStr, "api/") || strings.Contains(contentStr, "/v1/") ||
		strings.Contains(contentStr, "/v2/") || strings.Contains(contentStr, "/api") ||
		strings.Contains(contentStr, "endpoint") || strings.Contains(contentStr, "graphql") ||
		strings.Contains(contentStr, "rest") || strings.Contains(contentStr, "post ") ||
		strings.Contains(contentStr, "get ") || strings.Contains(contentStr, "put ") ||
		strings.Contains(contentStr, "delete "):
		// Treat API-like content as JavaScript for better extraction
		return JavaScript

	default:
		return Unknown
	}
}

func (a *Analyzer) generateContentHash(content []byte) string {
	// Use SHA256 for proper content hashing
	if len(content) == 0 {
		return "empty"
	}

	hasher := sha256.New()
	hasher.Write(content)
	hash := hasher.Sum(nil)

	// Return first 16 characters of hex representation for readability
	return hex.EncodeToString(hash)[:16]
}

func (a *Analyzer) isEmpty(finding *Finding) bool {
	return finding == nil ||
		(len(finding.Endpoints) == 0 &&
			len(finding.Parameters) == 0 &&
			len(finding.Headers) == 0 &&
			len(finding.Domains) == 0)
}

func (a *Analyzer) mergeFindings(findings []RawFinding, sourceURL string) *Finding {
	if len(findings) == 0 {
		return nil
	}

	merged := &Finding{
		Source: SourceInfo{
			FilePath: sourceURL,
			FileType: a.detectFileType(sourceURL, []byte{}).String(),
		},
	}

	// Aggregate all findings
	domainSet := make(map[string]bool)
	endpointSet := make(map[string]Endpoint)
	paramSet := make(map[string]Parameter)
	headerSet := make(map[string]Header)

	for _, finding := range findings {
		// Collect domains
		for _, domain := range finding.Domains {
			domainSet[domain] = true
		}

		// Collect endpoints (deduplicate by method+path)
		for _, ep := range finding.Endpoints {
			key := ep.Method + ":" + ep.Path
			endpointSet[key] = ep
		}

		// Collect parameters (deduplicate by name+type)
		for _, param := range finding.Parameters {
			key := param.Name + "_" + param.Type.String()
			paramSet[key] = param
		}

		// Collect headers (deduplicate by name)
		for _, header := range finding.Headers {
			headerSet[header.Name] = header
		}
	}

	// Convert sets to slices
	for domain := range domainSet {
		merged.Domains = append(merged.Domains, domain)
	}
	for _, ep := range endpointSet {
		merged.Endpoints = append(merged.Endpoints, ep)
	}
	for _, param := range paramSet {
		merged.Parameters = append(merged.Parameters, param)
	}
	for _, header := range headerSet {
		merged.Headers = append(merged.Headers, header)
	}

	return merged
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
