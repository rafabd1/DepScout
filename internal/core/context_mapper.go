package core

import (
	"strings"
)

/**
 * @description Context Mapper for analyzing relationships between endpoints,
 * parameters, and other findings based on proximity and semantic patterns
 */
type ContextMapper struct {
	proximityThreshold int
	semanticRules      map[string][]string
}

/**
 * @description Creates a new Context Mapper
 * @param proximityThreshold Maximum character distance for proximity analysis
 * @returns Initialized ContextMapper
 */
func NewContextMapper(proximityThreshold int) *ContextMapper {
	return &ContextMapper{
		proximityThreshold: proximityThreshold,
		semanticRules:      initializeSemanticRules(),
	}
}

/**
 * @description Maps context relationships in findings
 * @param findings Raw findings to analyze
 * @param content Original content for proximity analysis
 * @returns Enhanced findings with context relationships
 */
func (cm *ContextMapper) MapContext(findings []RawFinding, content []byte) []RawFinding {
	if len(findings) == 0 {
		return findings
	}

	contentStr := string(content)
	enhanced := make([]RawFinding, 0, len(findings))

	for _, finding := range findings {
		enhancedFinding := finding
		
		// Map endpoint-parameter relationships
		enhancedFinding.Endpoints = cm.mapEndpointContext(finding.Endpoints, finding.Parameters, contentStr)
		
		// Map parameter relationships
		enhancedFinding.Parameters = cm.mapParameterContext(finding.Parameters, contentStr)
		
		// Map header relationships
		enhancedFinding.Headers = cm.mapHeaderContext(finding.Headers, finding.Endpoints, contentStr)
		
		// Add semantic context
		enhancedFinding = cm.addSemanticContext(enhancedFinding, contentStr)
		
		enhanced = append(enhanced, enhancedFinding)
	}

	return enhanced
}

/**
 * @description Maps context for endpoints based on proximity to parameters
 * @param endpoints Endpoints to analyze
 * @param parameters Available parameters
 * @param content Content for proximity analysis
 * @returns Endpoints with mapped parameter relationships
 */
func (cm *ContextMapper) mapEndpointContext(endpoints []Endpoint, parameters []Parameter, content string) []Endpoint {
	enhanced := make([]Endpoint, 0, len(endpoints))

	for _, endpoint := range endpoints {
		enhancedEndpoint := endpoint
		enhancedEndpoint.RelatedParams = cm.findRelatedParameters(endpoint, parameters, content)
		enhancedEndpoint.Context = cm.enhanceEndpointContext(endpoint, content)
		enhanced = append(enhanced, enhancedEndpoint)
	}

	return enhanced
}

/**
 * @description Finds parameters related to an endpoint by proximity and semantics
 * @param endpoint Endpoint to analyze
 * @param parameters Available parameters
 * @param content Content for analysis
 * @returns Related parameter names
 */
func (cm *ContextMapper) findRelatedParameters(endpoint Endpoint, parameters []Parameter, content string) []string {
	related := make([]string, 0)
	endpointPos := strings.Index(content, endpoint.Path)
	
	if endpointPos == -1 {
		return related
	}

	// Find parameters by proximity
	for _, param := range parameters {
		paramPos := strings.Index(content, param.Name)
		if paramPos == -1 {
			continue
		}

		distance := abs(paramPos - endpointPos)
		if distance <= cm.proximityThreshold {
			related = append(related, param.Name)
		}
	}

	// Add semantically related parameters
	semanticParams := cm.findSemanticParams(endpoint.Path, parameters)
	for _, param := range semanticParams {
		if !contains(related, param) {
			related = append(related, param)
		}
	}

	return related
}

/**
 * @description Enhances endpoint context with additional metadata
 * @param endpoint Endpoint to enhance
 * @param content Content for analysis
 * @returns Enhanced context string
 */
func (cm *ContextMapper) enhanceEndpointContext(endpoint Endpoint, content string) string {
	context := endpoint.Context
	
	// Check for authentication context
	if cm.hasAuthContext(endpoint.Path, content) {
		context += "_auth"
	}
	
	// Check for admin context
	if cm.hasAdminContext(endpoint.Path, content) {
		context += "_admin"
	}
	
	// Check for API context
	if cm.hasAPIContext(endpoint.Path, content) {
		context += "_api"
	}

	return context
}

/**
 * @description Maps context for parameters based on usage patterns
 * @param parameters Parameters to analyze
 * @param content Content for analysis
 * @returns Parameters with enhanced context
 */
func (cm *ContextMapper) mapParameterContext(parameters []Parameter, content string) []Parameter {
	enhanced := make([]Parameter, 0, len(parameters))

	for _, param := range parameters {
		enhancedParam := param
		enhancedParam.Context = cm.enhanceParameterContext(param, content)
		enhancedParam.PossibleValues = cm.inferParameterValues(param, content)
		enhanced = append(enhanced, enhancedParam)
	}

	return enhanced
}

/**
 * @description Enhances parameter context with usage patterns
 * @param parameter Parameter to enhance
 * @param content Content for analysis
 * @returns Enhanced context string
 */
func (cm *ContextMapper) enhanceParameterContext(parameter Parameter, content string) string {
	context := parameter.Context
	
	// Check parameter type context
	if cm.isIDParameter(parameter.Name) {
		context += "_id"
	}
	
	if cm.isSearchParameter(parameter.Name) {
		context += "_search"
	}
	
	if cm.isPaginationParameter(parameter.Name) {
		context += "_pagination"
	}
	
	if cm.isFilterParameter(parameter.Name) {
		context += "_filter"
	}

	return context
}

/**
 * @description Infers possible values for parameters based on content
 * @param parameter Parameter to analyze
 * @param content Content for analysis
 * @returns Possible parameter values
 */
func (cm *ContextMapper) inferParameterValues(parameter Parameter, content string) []string {
	values := make([]string, 0)
	
	// Add common values based on parameter type
	if cm.isIDParameter(parameter.Name) {
		values = append(values, "1", "123", "user_id")
	}
	
	if cm.isSearchParameter(parameter.Name) {
		values = append(values, "test", "search_term", "query")
	}
	
	return values
}

/**
 * @description Maps context for headers based on endpoints
 * @param headers Headers to analyze
 * @param endpoints Available endpoints
 * @param content Content for analysis
 * @returns Headers with enhanced context
 */
func (cm *ContextMapper) mapHeaderContext(headers []Header, endpoints []Endpoint, content string) []Header {
	enhanced := make([]Header, 0, len(headers))

	for _, header := range headers {
		enhancedHeader := header
		enhancedHeader.Context = cm.enhanceHeaderContext(header, endpoints, content)
		enhancedHeader.RelatedEndpoints = cm.findRelatedEndpoints(header, endpoints, content)
		enhanced = append(enhanced, enhancedHeader)
	}

	return enhanced
}

/**
 * @description Enhances header context based on usage
 * @param header Header to enhance
 * @param endpoints Available endpoints
 * @param content Content for analysis
 * @returns Enhanced context string
 */
func (cm *ContextMapper) enhanceHeaderContext(header Header, endpoints []Endpoint, content string) string {
	context := header.Context
	
	if cm.isAuthHeader(header.Name) {
		context += "_auth"
	}
	
	if cm.isCORSHeader(header.Name) {
		context += "_cors"
	}
	
	if cm.isContentHeader(header.Name) {
		context += "_content"
	}

	return context
}

/**
 * @description Finds endpoints related to a header
 * @param header Header to analyze
 * @param endpoints Available endpoints
 * @param content Content for analysis
 * @returns Related endpoint paths
 */
func (cm *ContextMapper) findRelatedEndpoints(header Header, endpoints []Endpoint, content string) []string {
	related := make([]string, 0)
	headerPos := strings.Index(content, header.Name)
	
	if headerPos == -1 {
		return related
	}

	for _, endpoint := range endpoints {
		endpointPos := strings.Index(content, endpoint.Path)
		if endpointPos == -1 {
			continue
		}

		distance := abs(endpointPos - headerPos)
		if distance <= cm.proximityThreshold {
			related = append(related, endpoint.Path)
		}
	}

	return related
}

/**
 * @description Adds semantic context to findings based on patterns
 * @param finding Finding to enhance
 * @param content Content for analysis
 * @returns Finding with semantic context
 */
func (cm *ContextMapper) addSemanticContext(finding RawFinding, content string) RawFinding {
	enhanced := finding
	
	// Detect framework context
	if framework := cm.detectFramework(content); framework != "" {
		enhanced.Context += "_" + framework
	}
	
	// Detect security context
	if cm.hasSecurityContext(content) {
		enhanced.Context += "_security"
	}
	
	// Detect testing context
	if cm.hasTestingContext(content) {
		enhanced.Context += "_test"
	}

	return enhanced
}

// Helper methods for semantic analysis

/**
 * @description Finds semantically related parameters for an endpoint
 */
func (cm *ContextMapper) findSemanticParams(endpointPath string, parameters []Parameter) []string {
	related := make([]string, 0)
	
	// Extract path segments for analysis
	segments := strings.Split(strings.Trim(endpointPath, "/"), "/")
	
	for _, param := range parameters {
		for _, segment := range segments {
			if cm.areSemanticallySimilar(segment, param.Name) {
				related = append(related, param.Name)
				break
			}
		}
	}
	
	return related
}

/**
 * @description Checks if two strings are semantically similar
 */
func (cm *ContextMapper) areSemanticallySimilar(a, b string) bool {
	a, b = strings.ToLower(a), strings.ToLower(b)
	
	// Direct match
	if a == b {
		return true
	}
	
	// Substring match
	if strings.Contains(a, b) || strings.Contains(b, a) {
		return true
	}
	
	// Common variations
	variations := map[string][]string{
		"id":   {"identifier", "key", "uuid"},
		"user": {"username", "account", "profile"},
		"auth": {"authentication", "authorization", "token"},
	}
	
	for key, values := range variations {
		if (a == key || contains(values, a)) && (b == key || contains(values, b)) {
			return true
		}
	}
	
	return false
}

/**
 * @description Checks various context types
 */
func (cm *ContextMapper) hasAuthContext(path, content string) bool {
	authKeywords := []string{"auth", "login", "token", "bearer", "jwt"}
	return cm.hasKeywordsNear(path, authKeywords, content)
}

func (cm *ContextMapper) hasAdminContext(path, content string) bool {
	adminKeywords := []string{"admin", "dashboard", "management", "console"}
	return cm.hasKeywordsNear(path, adminKeywords, content) || strings.Contains(path, "/admin")
}

func (cm *ContextMapper) hasAPIContext(path, content string) bool {
	return strings.Contains(path, "/api") || strings.Contains(path, "/v")
}

func (cm *ContextMapper) hasSecurityContext(content string) bool {
	securityKeywords := []string{"csrf", "xss", "cors", "security", "sanitize"}
	for _, keyword := range securityKeywords {
		if strings.Contains(strings.ToLower(content), keyword) {
			return true
		}
	}
	return false
}

func (cm *ContextMapper) hasTestingContext(content string) bool {
	testKeywords := []string{"test", "mock", "stub", "spec", "jest", "mocha"}
	for _, keyword := range testKeywords {
		if strings.Contains(strings.ToLower(content), keyword) {
			return true
		}
	}
	return false
}

func (cm *ContextMapper) detectFramework(content string) string {
	frameworks := map[string][]string{
		"react":   {"React", "useState", "useEffect", "jsx"},
		"angular": {"@angular", "NgModule", "Component"},
		"vue":     {"Vue", "createApp", "ref", "reactive"},
		"express": {"express", "app.get", "router."},
		"jquery":  {"$", "jQuery", ".ajax"},
	}
	
	for framework, keywords := range frameworks {
		for _, keyword := range keywords {
			if strings.Contains(content, keyword) {
				return framework
			}
		}
	}
	
	return ""
}

/**
 * @description Parameter type checks
 */
func (cm *ContextMapper) isIDParameter(name string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, "id") || name == "key" || name == "uuid"
}

func (cm *ContextMapper) isSearchParameter(name string) bool {
	name = strings.ToLower(name)
	searchParams := []string{"q", "query", "search", "term", "keyword"}
	return contains(searchParams, name)
}

func (cm *ContextMapper) isPaginationParameter(name string) bool {
	name = strings.ToLower(name)
	paginationParams := []string{"page", "limit", "offset", "size", "count", "per_page"}
	return contains(paginationParams, name)
}

func (cm *ContextMapper) isFilterParameter(name string) bool {
	name = strings.ToLower(name)
	filterParams := []string{"filter", "sort", "order", "by", "category", "status", "type"}
	return contains(filterParams, name)
}

/**
 * @description Header type checks
 */
func (cm *ContextMapper) isAuthHeader(name string) bool {
	name = strings.ToLower(name)
	authHeaders := []string{"authorization", "x-auth-token", "x-api-key", "bearer"}
	return contains(authHeaders, name) || strings.Contains(name, "auth")
}

func (cm *ContextMapper) isCORSHeader(name string) bool {
	name = strings.ToLower(name)
	return strings.HasPrefix(name, "access-control-") || name == "origin"
}

func (cm *ContextMapper) isContentHeader(name string) bool {
	name = strings.ToLower(name)
	contentHeaders := []string{"content-type", "accept", "content-length", "content-encoding"}
	return contains(contentHeaders, name)
}

/**
 * @description Checks if keywords appear near a target string in content
 */
func (cm *ContextMapper) hasKeywordsNear(target string, keywords []string, content string) bool {
	targetPos := strings.Index(strings.ToLower(content), strings.ToLower(target))
	if targetPos == -1 {
		return false
	}
	
	start := max(0, targetPos-cm.proximityThreshold)
	end := min(len(content), targetPos+len(target)+cm.proximityThreshold)
	nearby := strings.ToLower(content[start:end])
	
	for _, keyword := range keywords {
		if strings.Contains(nearby, strings.ToLower(keyword)) {
			return true
		}
	}
	
	return false
}

/**
 * @description Initializes semantic rules for context mapping
 */
func initializeSemanticRules() map[string][]string {
	return map[string][]string{
		"authentication": {"auth", "login", "token", "bearer", "jwt", "session"},
		"authorization":  {"admin", "role", "permission", "access", "privilege"},
		"user_management": {"user", "profile", "account", "username", "email"},
		"data_operations": {"create", "read", "update", "delete", "crud"},
		"search": {"search", "query", "find", "filter", "sort"},
		"pagination": {"page", "limit", "offset", "size", "count"},
		"upload": {"upload", "file", "media", "attachment", "image"},
		"api": {"api", "endpoint", "service", "resource", "rest"},
	}
}

// Utility functions

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}