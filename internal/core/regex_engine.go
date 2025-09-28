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
				
				// Validate parameter before including
				if !re.looksLikeParameter(paramName) {
					continue
				}
				
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
	if len(path) < 2 || len(path) > 200 {
		return false
	}

	// Must start with / for API endpoints
	if !strings.HasPrefix(path, "/") {
		return false
	}

	// Skip obvious false positives - static resources
	staticPaths := []string{"/", "//", "/css", "/js", "/img", "/images", "/static", "/assets", "/public"}
	for _, staticPath := range staticPaths {
		if path == staticPath || strings.HasPrefix(path, staticPath+"/") {
			return false
		}
	}

	// Skip JavaScript code patterns
	codePatterns := []string{
		"${", "\\", "/*", "*/", "//", "function", "return", "var ", "let ", "const ", 
		"if(", "for(", "while(", ".map(", ".join(", ".replace(", ".split(",
		"&&", "||", "===", "!==", "==", "!=", "++", "--",
	}
	for _, codePattern := range codePatterns {
		if strings.Contains(path, codePattern) {
			return false
		}
	}

	// Skip paths with too many special characters or numbers
	specialCount := 0
	digitCount := 0
	for _, char := range path {
		if char == '{' || char == '}' || char == '$' || char == '(' || char == ')' || char == '[' || char == ']' {
			specialCount++
		}
		if char >= '0' && char <= '9' {
			digitCount++
		}
	}
	
	// Reject if too many special characters or mostly numbers
	if specialCount > 3 || digitCount > len(path)/2 {
		return false
	}

	// Must contain at least one letter for valid API paths
	hasLetter := false
	for _, char := range path {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetter = true
			break
		}
	}
	
	if !hasLetter {
		return false
	}

	// Validate API-like structure - should have reasonable segments
	segments := strings.Split(path, "/")
	if len(segments) > 10 { // Too many segments likely not an API
		return false
	}

	// Check for common API patterns
	for _, segment := range segments {
		if segment == "api" || segment == "v1" || segment == "v2" || segment == "v3" || 
		   segment == "auth" || segment == "admin" || segment == "users" || segment == "user" ||
		   strings.HasPrefix(segment, "api") {
			return true
		}
	}

	// Allow simple RESTful patterns
	if len(segments) >= 2 && len(segments) <= 6 {
		return true
	}

	return false
}

func (re *RegexEngine) looksLikeParameter(param string) bool {
	if len(param) < 2 || len(param) > 50 {
		return false
	}

	// Basic validation - must contain letters and be reasonable length
	if !re.hasValidCharacters(param) {
		return false
	}

	// Use contextual analysis instead of blocklists
	return re.looksLikeAPIParameter(param)
}

func (re *RegexEngine) hasValidCharacters(param string) bool {
	// Must contain letters
	hasLetter := false
	
	for _, char := range param {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetter = true
		}
		
		// Only reject obvious code patterns that can't be parameters
		if char == '{' || char == '}' || char == '(' || char == ')' || 
		   char == '[' || char == ']' || char == '=' || char == '!' {
			return false
		}
	}
	
	return hasLetter
}

func (re *RegexEngine) looksLikeAPIParameter(param string) bool {
	paramLower := strings.ToLower(param)
	
	// 1. Semantic Analysis - API parameters follow specific patterns
	if re.hasAPISemantics(paramLower) {
		return true
	}
	
	// 2. Structural Analysis - API naming conventions
	if re.hasAPIStructure(param) {
		return true
	}
	
	// 3. Reject obvious function/method patterns
	if re.looksLikeFunction(param) {
		return false
	}
	
	// 4. Reject object property patterns that aren't API-like
	if re.looksLikeObjectProperty(param) {
		return false
	}
	
	// 5. Only accept very conservative API patterns
	return re.passesStrictAPIHeuristics(param)
}

func (re *RegexEngine) hasAPISemantics(paramLower string) bool {
	// Core API parameter concepts
	apiConcepts := []string{
		"id", "key", "token", "auth", "session", "csrf",
		"user", "email", "password", "name", "title", 
		"page", "limit", "offset", "sort", "order", "filter", "search", "query",
		"category", "tag", "status", "type", "format", "version",
		"date", "time", "created", "updated", "from", "to",
		"price", "amount", "count", "total", "size",
	}
	
	for _, concept := range apiConcepts {
		if paramLower == concept || paramLower == concept+"s" ||
		   strings.HasSuffix(paramLower, "_"+concept) || strings.HasSuffix(paramLower, concept+"_id") {
			return true
		}
	}
	
	return false
}

func (re *RegexEngine) hasAPIStructure(param string) bool {
	// Snake_case with API-like components
	if strings.Contains(param, "_") {
		parts := strings.Split(strings.ToLower(param), "_")
		if len(parts) >= 2 {
			// Check if parts contain API-like terms
			for _, part := range parts {
				if len(part) >= 2 && (part == "id" || part == "key" || part == "name" || 
					part == "type" || part == "url" || part == "api" || part == "user") {
					return true
				}
			}
		}
	}
	
	// CamelCase with specific API patterns (not function-like)
	if !strings.Contains(param, "_") && param != strings.ToLower(param) {
		// Should be noun-like, not verb-like for API parameters
		if !re.startsWithVerb(strings.ToLower(param)) {
			return len(param) >= 4 && len(param) <= 25
		}
	}
	
	return false
}

func (re *RegexEngine) looksLikeFunction(param string) bool {
	paramLower := strings.ToLower(param)
	
	// Only reject very obvious function patterns
	obviousFunctions := []string{"canParse", "canRead", "canWrite", "canDereference", 
		"getData", "setData", "createElement", "addEventListener", "removeEventListener"}
	
	for _, fn := range obviousFunctions {
		if strings.EqualFold(param, fn) {
			return true
		}
	}
	
	// Reject camelCase that starts with obvious verbs and is long enough to be a function
	if len(param) > 8 {
		verbPrefixes := []string{"get", "set", "create", "update", "delete", "validate", "parse", "render", "handle"}
		for _, verb := range verbPrefixes {
			if strings.HasPrefix(paramLower, verb) && len(param) > len(verb) {
				// Check if next char is uppercase (camelCase function pattern)
				if param[len(verb)] >= 'A' && param[len(verb)] <= 'Z' {
					return true
				}
			}
		}
	}
	
	return false
}

func (re *RegexEngine) looksLikeObjectProperty(param string) bool {
	paramLower := strings.ToLower(param)
	
	// JSON Schema specific properties
	jsonSchemaProps := []string{"schema", "vocabulary", "anchor", "dynamicanchor", "dynamicref", 
		"definitions", "properties", "patternproperties", "additionalproperties"}
	for _, prop := range jsonSchemaProps {
		if paramLower == prop || strings.Contains(paramLower, prop) {
			return true
		}
	}
	
	// React/Framework properties
	frameworkProps := []string{"component", "element", "classname", "onclick", "onchange", 
		"props", "state", "context", "provider", "consumer"}
	for _, prop := range frameworkProps {
		if paramLower == prop || strings.Contains(paramLower, prop) {
			return true
		}
	}
	
	// General JavaScript object patterns
	jsPatterns := []string{"typeof", "instanceof", "constructor", "prototype", "visitor", 
		"iterator", "descriptor", "accessor", "invariant"}
	for _, pattern := range jsPatterns {
		if paramLower == pattern {
			return true
		}
	}
	
	return false
}

func (re *RegexEngine) startsWithVerb(paramLower string) bool {
	verbs := []string{"can", "is", "has", "get", "set", "add", "remove", "delete", 
		"create", "update", "check", "validate", "parse", "render", "handle", "should", "will"}
	
	for _, verb := range verbs {
		if strings.HasPrefix(paramLower, verb) {
			return true
		}
	}
	
	return false
}

func (re *RegexEngine) passesStrictAPIHeuristics(param string) bool {
	// Ultra-conservative - only accept parameters with high API confidence
	paramLower := strings.ToLower(param)
	
	// Core API parameter names (very common in real APIs)
	coreAPIParams := []string{"id", "key", "token", "auth", "user", "email", "page", 
		"limit", "offset", "sort", "order", "search", "query", "filter", "type", "format", 
		"category", "status", "name", "url", "code", "version"}
	
	for _, coreParam := range coreAPIParams {
		if paramLower == coreParam {
			return true
		}
	}
	
	// Parameters ending with API suffixes
	if len(param) >= 4 && len(param) <= 20 {
		if strings.HasSuffix(paramLower, "id") || strings.HasSuffix(paramLower, "key") ||
		   strings.HasSuffix(paramLower, "url") || strings.HasSuffix(paramLower, "code") {
			return true
		}
	}
	
	// Snake_case parameters (common in APIs)
	if strings.Contains(param, "_") && len(param) >= 5 && len(param) <= 25 {
		parts := strings.Split(paramLower, "_")
		for _, part := range parts {
			if part == "id" || part == "key" || part == "url" || part == "api" {
				return true
			}
		}
	}
	
	return false
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