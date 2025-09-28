package core

import (
	"regexp"
	"strings"
)

/**
 * @description High-performance regex engine optimized for Go with pre-compiled patterns
 * for fast endpoint, parameter, header, and domain extraction
 */
type RegexEngine struct {
	// Pre-compiled patterns for optimal performance
	endpointPatterns   []*regexp.Regexp
	parameterPatterns  []*regexp.Regexp
	headerPatterns     []*regexp.Regexp
	domainPatterns     []*regexp.Regexp
	methodPatterns     []*regexp.Regexp
	bodyPatterns       []*regexp.Regexp
	htmlPatterns       []*regexp.Regexp
	jsonPatterns       []*regexp.Regexp
}

/**
 * @description Creates a new RegexEngine with all patterns pre-compiled for performance
 * @returns Configured RegexEngine ready for extraction
 */
func NewRegexEngine() *RegexEngine {
	return &RegexEngine{
		endpointPatterns: compileEndpointPatterns(),
		parameterPatterns: compileParameterPatterns(),
		headerPatterns: compileHeaderPatterns(),
		domainPatterns: compileDomainPatterns(),
		methodPatterns: compileMethodPatterns(),
		bodyPatterns: compileBodyPatterns(),
		htmlPatterns: compileHTMLPatterns(),
		jsonPatterns: compileJSONPatterns(),
	}
}

/**
 * @description Main extraction method that combines all regex extraction techniques
 * @param content Content bytes to analyze
 * @param sourceURL Source identifier for context
 * @returns Array of raw findings from regex analysis
 */
func (re *RegexEngine) ExtractAll(content []byte, sourceURL string) []RawFinding {
	contentStr := string(content)
	var findings []RawFinding

	// Extract different types of data
	endpoints := re.extractEndpoints(contentStr)
	parameters := re.extractParameters(contentStr)
	headers := re.extractHeaders(contentStr)
	domains := re.extractDomains(contentStr)
	bodies := re.extractBodies(contentStr)

	if len(endpoints) > 0 || len(parameters) > 0 || len(headers) > 0 || len(domains) > 0 {
		finding := RawFinding{
			Endpoints:  endpoints,
			Parameters: parameters,
			Headers:    headers,
			Domains:    domains,
			Bodies:     bodies,
			Context:    "regex_extraction",
			Confidence: 0.7, // Medium-high confidence for regex
		}
		findings = append(findings, finding)
	}

	return findings
}

/**
 * @description Specialized HTML extraction focusing on script tags, forms, AJAX calls
 * @param content HTML content bytes
 * @param sourceURL Source identifier
 * @returns Raw findings from HTML analysis
 */
func (re *RegexEngine) ExtractFromHTML(content []byte, sourceURL string) []RawFinding {
	contentStr := string(content)
	var findings []RawFinding

	// Extract from script tags
	scriptEndpoints := re.extractFromScriptTags(contentStr)
	
	// Extract from form actions
	formEndpoints := re.extractFromForms(contentStr)
	
	// Extract from AJAX calls
	ajaxEndpoints := re.extractFromAJAX(contentStr)

	// Combine HTML-specific findings
	allEndpoints := append(scriptEndpoints, formEndpoints...)
	allEndpoints = append(allEndpoints, ajaxEndpoints...)

	if len(allEndpoints) > 0 {
		finding := RawFinding{
			Endpoints:  allEndpoints,
			Context:    "html_extraction",
			Confidence: 0.8, // High confidence for HTML structures
		}
		findings = append(findings, finding)
	}

	return findings
}

/**
 * @description Specialized JSON extraction for configuration files and API definitions
 * @param content JSON content bytes
 * @param sourceURL Source identifier  
 * @returns Raw findings from JSON analysis
 */
func (re *RegexEngine) ExtractFromJSON(content []byte, sourceURL string) []RawFinding {
	contentStr := string(content)
	var findings []RawFinding

	// Extract API endpoints from JSON configs
	endpoints := re.extractJSONEndpoints(contentStr)
	
	// Extract parameter definitions
	parameters := re.extractJSONParameters(contentStr)

	if len(endpoints) > 0 || len(parameters) > 0 {
		finding := RawFinding{
			Endpoints:  endpoints,
			Parameters: parameters,
			Context:    "json_extraction",
			Confidence: 0.9, // Very high confidence for structured JSON
		}
		findings = append(findings, finding)
	}

	return findings
}

/**
 * @description Minimal URL pattern extraction as last resort
 * @param content Content bytes to scan
 * @returns Basic URL patterns found
 */
func (re *RegexEngine) ExtractURLPatterns(content []byte) []string {
	contentStr := string(content)
	var urls []string

	// Basic URL pattern - very permissive
	urlPattern := regexp.MustCompile(`/[a-zA-Z0-9/_\-\.]+`)
	matches := urlPattern.FindAllString(contentStr, -1)

	for _, match := range matches {
		if len(match) > 3 && re.looksLikeEndpoint(match) {
			urls = append(urls, match)
		}
	}

	return re.deduplicateStrings(urls)
}

// Private extraction methods

func (re *RegexEngine) extractEndpoints(content string) []Endpoint {
	var endpoints []Endpoint
	seen := make(map[string]bool)

	for _, pattern := range re.endpointPatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				path := match[1]
				if seen[path] || !re.looksLikeEndpoint(path) {
					continue
				}
				seen[path] = true

				endpoint := Endpoint{
					Path:    path,
					Method:  re.inferMethodFromContext(path, content),
					Context: re.getContext(content, match[0]),
				}
				endpoints = append(endpoints, endpoint)
			}
		}
	}

	return endpoints
}

func (re *RegexEngine) extractParameters(content string) []Parameter {
	var parameters []Parameter
	seen := make(map[string]bool)

	for _, pattern := range re.parameterPatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				paramName := match[1]
				key := paramName + "_url" // Default to URL param
				if seen[key] {
					continue
				}
				seen[key] = true

				param := Parameter{
					Name:     paramName,
					Type:     URLParam, // Default, can be enhanced by AST
					Context:  re.getContext(content, match[0]),
					DataType: "string", // Default
				}
				parameters = append(parameters, param)
			}
		}
	}

	return parameters
}

func (re *RegexEngine) extractHeaders(content string) []Header {
	var headers []Header
	seen := make(map[string]bool)

	for _, pattern := range re.headerPatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				headerName := match[1]
				if seen[headerName] {
					continue
				}
				seen[headerName] = true

				valuePattern := ""
				if len(match) > 2 {
					valuePattern = match[2]
				}

				header := Header{
					Name:         headerName,
					ValuePattern: valuePattern,
					Context:      re.getContext(content, match[0]),
				}
				headers = append(headers, header)
			}
		}
	}

	return headers
}

func (re *RegexEngine) extractDomains(content string) []string {
	var domains []string
	seen := make(map[string]bool)

	for _, pattern := range re.domainPatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				domain := match[1]
				if seen[domain] {
					continue
				}
				seen[domain] = true
				domains = append(domains, domain)
			}
		}
	}

	return domains
}

func (re *RegexEngine) extractBodies(content string) []RequestBody {
	var bodies []RequestBody

	for _, pattern := range re.bodyPatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				body := RequestBody{
					ContentType: re.inferContentType(match[1]),
					Structure:   match[1],
					Context:     re.getContext(content, match[0]),
				}
				bodies = append(bodies, body)
			}
		}
	}

	return bodies
}

// HTML-specific extraction methods

func (re *RegexEngine) extractFromScriptTags(content string) []Endpoint {
	var endpoints []Endpoint
	
	for _, pattern := range re.htmlPatterns {
		if strings.Contains(pattern.String(), "script") {
			matches := pattern.FindAllStringSubmatch(content, -1)
			for _, match := range matches {
				if len(match) > 1 && re.looksLikeEndpoint(match[1]) {
					endpoint := Endpoint{
						Path:    match[1],
						Method:  "GET", // Default for script sources
						Context: "script_tag",
					}
					endpoints = append(endpoints, endpoint)
				}
			}
		}
	}
	
	return endpoints
}

func (re *RegexEngine) extractFromForms(content string) []Endpoint {
	var endpoints []Endpoint
	
	// Form action pattern
	formPattern := regexp.MustCompile(`<form[^>]*action\s*=\s*["']([^"']+)["']`)
	matches := formPattern.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches {
		if len(match) > 1 && re.looksLikeEndpoint(match[1]) {
			endpoint := Endpoint{
				Path:    match[1],
				Method:  "POST", // Default for forms
				Context: "form_action",
			}
			endpoints = append(endpoints, endpoint)
		}
	}
	
	return endpoints
}

func (re *RegexEngine) extractFromAJAX(content string) []Endpoint {
	var endpoints []Endpoint
	
	// jQuery AJAX pattern
	ajaxPattern := regexp.MustCompile(`\$\.(?:ajax|get|post)\s*\(\s*["']([^"']+)["']`)
	matches := ajaxPattern.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches {
		if len(match) > 1 && re.looksLikeEndpoint(match[1]) {
			endpoint := Endpoint{
				Path:    match[1],
				Method:  re.inferMethodFromAjax(match[0]),
				Context: "ajax_call",
			}
			endpoints = append(endpoints, endpoint)
		}
	}
	
	return endpoints
}

// JSON-specific extraction methods

func (re *RegexEngine) extractJSONEndpoints(content string) []Endpoint {
	var endpoints []Endpoint
	
	for _, pattern := range re.jsonPatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 && re.looksLikeEndpoint(match[1]) {
				endpoint := Endpoint{
					Path:    match[1],
					Method:  "GET", // Default
					Context: "json_config",
				}
				endpoints = append(endpoints, endpoint)
			}
		}
	}
	
	return endpoints
}

func (re *RegexEngine) extractJSONParameters(content string) []Parameter {
	var parameters []Parameter
	
	// JSON parameter pattern
	paramPattern := regexp.MustCompile(`"(\w+)"\s*:\s*"[^"]*"`)
	matches := paramPattern.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			param := Parameter{
				Name:     match[1],
				Type:     BodyParam, // Assume JSON body params
				Context:  "json_property",
				DataType: "string",
			}
			parameters = append(parameters, param)
		}
	}
	
	return parameters
}

// Helper methods

func (re *RegexEngine) inferMethodFromContext(path, content string) string {
	context := re.getContextWindow(content, path, 100)
	contextLower := strings.ToLower(context)

	// Look for method indicators in proximity
	for _, pattern := range re.methodPatterns {
		if matches := pattern.FindStringSubmatch(context); len(matches) > 1 {
			return strings.ToUpper(matches[1])
		}
	}

	// Fallback inference based on common patterns
	switch {
	case strings.Contains(contextLower, "post") || strings.Contains(contextLower, "create"):
		return "POST"
	case strings.Contains(contextLower, "put") || strings.Contains(contextLower, "update"):
		return "PUT"
	case strings.Contains(contextLower, "delete") || strings.Contains(contextLower, "remove"):
		return "DELETE"
	default:
		return "GET"
	}
}

func (re *RegexEngine) inferMethodFromAjax(ajaxCall string) string {
	if strings.Contains(strings.ToLower(ajaxCall), ".post") {
		return "POST"
	}
	return "GET"
}

func (re *RegexEngine) inferContentType(body string) string {
	body = strings.TrimSpace(body)
	if strings.HasPrefix(body, "{") && strings.HasSuffix(body, "}") {
		return "application/json"
	}
	if strings.Contains(body, "=") && strings.Contains(body, "&") {
		return "application/x-www-form-urlencoded"
	}
	return "text/plain"
}

func (re *RegexEngine) getContext(content, match string) string {
	return re.getContextWindow(content, match, 50)
}

func (re *RegexEngine) getContextWindow(content, match string, windowSize int) string {
	index := strings.Index(content, match)
	if index == -1 {
		return match
	}

	start := max(0, index-windowSize)
	end := min(len(content), index+len(match)+windowSize)
	
	return content[start:end]
}

func (re *RegexEngine) looksLikeEndpoint(path string) bool {
	if len(path) < 2 {
		return false
	}

	// Must start with / or contain ://
	if !strings.HasPrefix(path, "/") && !strings.Contains(path, "://") {
		return false
	}

	// Skip obvious false positives
	falsePaths := []string{"/", "//", "/css", "/js", "/img", "/images", "/static"}
	for _, falseP := range falsePaths {
		if path == falseP {
			return false
		}
	}

	return true
}

func (re *RegexEngine) deduplicateStrings(strs []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, str := range strs {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}
	
	return result
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Pattern compilation functions - defined in separate file for organization