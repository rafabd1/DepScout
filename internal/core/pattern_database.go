package core

import (
	"strings"
)

/**
 * @description Pattern Database containing curated extraction patterns and rules
 * for different frameworks, libraries, and common web application structures
 */
type PatternDatabase struct {
	frameworkPatterns map[string][]string
	commonPatterns    []string
	customPatterns    []string
}

/**
 * @description Creates a new Pattern Database with predefined patterns
 * @returns Initialized PatternDatabase with common web patterns
 */
func NewPatternDatabase() *PatternDatabase {
	return &PatternDatabase{
		frameworkPatterns: initializeFrameworkPatterns(),
		commonPatterns:    initializeCommonPatterns(),
		customPatterns:    make([]string, 0),
	}
}

/**
 * @description Gets patterns for a specific framework or library
 * @param framework Framework name (e.g., "express", "react", "angular")
 * @returns Array of framework-specific patterns
 */
func (pd *PatternDatabase) GetFrameworkPatterns(framework string) []string {
	if patterns, exists := pd.frameworkPatterns[framework]; exists {
		return patterns
	}
	return []string{}
}

/**
 * @description Gets all common patterns regardless of framework
 * @returns Array of common extraction patterns
 */
func (pd *PatternDatabase) GetCommonPatterns() []string {
	return pd.commonPatterns
}

/**
 * @description Adds custom patterns for specific use cases
 * @param patterns Custom patterns to add
 */
func (pd *PatternDatabase) AddCustomPatterns(patterns []string) {
	pd.customPatterns = append(pd.customPatterns, patterns...)
}

/**
 * @description Gets all patterns combined (common + framework + custom)
 * @param framework Optional framework name for framework-specific patterns
 * @returns Combined array of all relevant patterns
 */
func (pd *PatternDatabase) GetAllPatterns(framework string) []string {
	all := make([]string, 0)
	
	// Add common patterns
	all = append(all, pd.commonPatterns...)
	
	// Add framework-specific patterns if specified
	if framework != "" {
		all = append(all, pd.GetFrameworkPatterns(framework)...)
	}
	
	// Add custom patterns
	all = append(all, pd.customPatterns...)
	
	return all
}

/**
 * @description Initializes framework-specific patterns
 * @returns Map of framework names to their patterns
 */
func initializeFrameworkPatterns() map[string][]string {
	return map[string][]string{
		"express": {
			// Express.js route definitions
			`app\.(get|post|put|delete|patch)\s*\(\s*["'\x60]([/][^"'\x60]*)["'\x60']`,
			`router\.(get|post|put|delete|patch)\s*\(\s*["'\x60]([/][^"'\x60]*)["'\x60']`,
			`\.route\s*\(\s*["'\x60]([/][^"'\x60]*)["'\x60']\)`,
		},
		"react": {
			// React Router patterns
			`<Route\s+path\s*=\s*["'\x60]([/][^"'\x60]*)["'\x60']`,
			`useHistory\(\)\.push\s*\(\s*["'\x60]([/][^"'\x60]*)["'\x60']`,
			`history\.push\s*\(\s*["'\x60]([/][^"'\x60]*)["'\x60']`,
			`Link\s+to\s*=\s*["'\x60]([/][^"'\x60]*)["'\x60']`,
		},
		"angular": {
			// Angular routing patterns
			`path\s*:\s*["'\x60]([^"'\x60]*)["'\x60']`,
			`routerLink\s*=\s*["'\x60]([/][^"'\x60]*)["'\x60']`,
			`router\.navigate\s*\(\s*\[\s*["'\x60]([/][^"'\x60]*)["'\x60']`,
		},
		"vue": {
			// Vue.js routing patterns
			`path\s*:\s*["'\x60]([/][^"'\x60]*)["'\x60']`,
			`router-link\s+to\s*=\s*["'\x60]([/][^"'\x60]*)["'\x60']`,
			`\$router\.push\s*\(\s*["'\x60]([/][^"'\x60]*)["'\x60']`,
		},
		"django": {
			// Django URL patterns (in JavaScript context)
			`url\s*:\s*["'\x60]([/][^"'\x60]*)["'\x60']`,
			`reverse\s*\(\s*["'\x60]([^"'\x60]*)["'\x60']`,
		},
		"spring": {
			// Spring Boot REST endpoints (in JavaScript context)
			`RequestMapping\s*\(\s*["'\x60]([/][^"'\x60]*)["'\x60']`,
			`GetMapping\s*\(\s*["'\x60]([/][^"'\x60]*)["'\x60']`,
			`PostMapping\s*\(\s*["'\x60]([/][^"'\x60]*)["'\x60']`,
		},
		"laravel": {
			// Laravel route patterns (in JavaScript context)
			`route\s*\(\s*["'\x60]([^"'\x60]*)["'\x60']`,
			`action\s*\(\s*["'\x60]([^"'\x60]*)["'\x60']`,
		},
		"wordpress": {
			// WordPress-specific patterns
			`admin_url\s*\(\s*["'\x60]([^"'\x60]*)["'\x60']`,
			`wp_ajax_[a-zA-Z0-9_]+`,
			`ajaxurl`,
		},
	}
}

/**
 * @description Initializes common patterns used across frameworks
 * @returns Array of common extraction patterns
 */
func initializeCommonPatterns() []string {
	return []string{
		// API endpoint patterns
		`["'\x60](/api[/a-zA-Z0-9_\-]*)["\x60']`,
		`["'\x60](/v\d+[/a-zA-Z0-9_\-]*)["\x60']`,
		
		// Admin patterns
		`["'\x60](/admin[/a-zA-Z0-9_\-]*)["\x60']`,
		`["'\x60](/dashboard[/a-zA-Z0-9_\-]*)["\x60']`,
		
		// Authentication patterns
		`["'\x60](/auth[/a-zA-Z0-9_\-]*)["\x60']`,
		`["'\x60](/login[/a-zA-Z0-9_\-]*)["\x60']`,
		`["'\x60](/logout[/a-zA-Z0-9_\-]*)["\x60']`,
		`["'\x60](/register[/a-zA-Z0-9_\-]*)["\x60']`,
		
		// User management patterns
		`["'\x60](/users?[/a-zA-Z0-9_\-]*)["\x60']`,
		`["'\x60](/profile[/a-zA-Z0-9_\-]*)["\x60']`,
		`["'\x60](/account[/a-zA-Z0-9_\-]*)["\x60']`,
		
		// Resource patterns
		`["'\x60](/upload[/a-zA-Z0-9_\-]*)["\x60']`,
		`["'\x60](/download[/a-zA-Z0-9_\-]*)["\x60']`,
		`["'\x60](/files?[/a-zA-Z0-9_\-]*)["\x60']`,
		
		// Search patterns
		`["'\x60](/search[/a-zA-Z0-9_\-]*)["\x60']`,
		`["'\x60](/query[/a-zA-Z0-9_\-]*)["\x60']`,
		
		// WebSocket patterns
		`["'\x60](/ws[/a-zA-Z0-9_\-]*)["\x60']`,
		`["'\x60](/socket\.io[/a-zA-Z0-9_\-]*)["\x60']`,
		
		// GraphQL patterns
		`["'\x60](/graphql[/]?)["\x60']`,
		
		// Health check patterns
		`["'\x60](/health[/a-zA-Z0-9_\-]*)["\x60']`,
		`["'\x60](/status[/a-zA-Z0-9_\-]*)["\x60']`,
		
		// Configuration patterns
		`["'\x60](/config[/a-zA-Z0-9_\-]*)["\x60']`,
		`["'\x60](/settings[/a-zA-Z0-9_\-]*)["\x60']`,
	}
}

/**
 * @description Gets parameter patterns for common parameter types
 * @returns Map of parameter types to their patterns
 */
func (pd *PatternDatabase) GetParameterPatterns() map[string][]string {
	return map[string][]string{
		"id_params": {
			`\{id\}`, `\{[a-zA-Z]*[Ii]d\}`, `:id`, `:.*[Ii]d`,
		},
		"auth_params": {
			`token`, `auth`, `authorization`, `bearer`, `api[_-]?key`,
		},
		"pagination": {
			`page`, `limit`, `offset`, `size`, `count`, `per[_-]?page`,
		},
		"search": {
			`q`, `query`, `search`, `term`, `keyword`,
		},
		"filters": {
			`filter`, `sort`, `order`, `by`, `category`, `status`,
		},
		"format": {
			`format`, `type`, `content[_-]?type`, `accept`,
		},
	}
}

/**
 * @description Gets header patterns for common HTTP headers
 * @returns Map of header categories to their patterns
 */
func (pd *PatternDatabase) GetHeaderPatterns() map[string][]string {
	return map[string][]string{
		"auth_headers": {
			`Authorization`, `Bearer`, `X-Auth-Token`, `X-API-Key`, `X-Access-Token`,
		},
		"content_headers": {
			`Content-Type`, `Accept`, `Content-Length`, `Content-Encoding`,
		},
		"cors_headers": {
			`Access-Control-[A-Za-z-]+`, `Origin`, `X-Requested-With`,
		},
		"custom_headers": {
			`X-[A-Za-z-]+`, `[A-Za-z]+-[A-Za-z-]+-[A-Za-z-]+`,
		},
		"security_headers": {
			`X-CSRF-Token`, `X-Frame-Options`, `X-Content-Type-Options`,
		},
		"cache_headers": {
			`Cache-Control`, `ETag`, `If-None-Match`, `Last-Modified`,
		},
	}
}

/**
 * @description Gets domain patterns for common service types
 * @returns Map of service types to their domain patterns
 */
func (pd *PatternDatabase) GetDomainPatterns() map[string][]string {
	return map[string][]string{
		"api_domains": {
			`api\.`, `service\.`, `gateway\.`, `backend\.`,
		},
		"admin_domains": {
			`admin\.`, `dashboard\.`, `management\.`, `console\.`,
		},
		"cdn_domains": {
			`cdn\.`, `static\.`, `assets\.`, `media\.`,
		},
		"dev_domains": {
			`dev\.`, `test\.`, `staging\.`, `local\.`,
		},
		"internal_domains": {
			`internal\.`, `corp\.`, `intranet\.`, `private\.`,
		},
	}
}

/**
 * @description Detects framework based on content patterns
 * @param content Content to analyze
 * @returns Detected framework name or empty string
 */
func (pd *PatternDatabase) DetectFramework(content string) string {
	// Framework detection patterns
	frameworks := map[string][]string{
		"react": {"React", "ReactDOM", "useState", "useEffect", "jsx"},
		"angular": {"@angular", "NgModule", "Component", "Injectable"},
		"vue": {"Vue", "createApp", "ref", "reactive", "v-"},
		"express": {"express", "app.get", "app.post", "router."},
		"jquery": {"$", "jQuery", ".ajax", ".get", ".post"},
		"axios": {"axios", "axios.get", "axios.post"},
	}

	for framework, patterns := range frameworks {
		for _, pattern := range patterns {
			if strings.Contains(content, pattern) {
				return framework
			}
		}
	}

	return ""
}

/**
 * @description Gets confidence level for a pattern match
 * @param pattern Matched pattern
 * @param context Context where pattern was found
 * @returns Confidence score between 0.0 and 1.0
 */
func (pd *PatternDatabase) GetPatternConfidence(pattern, context string) float64 {
	// High confidence patterns
	highConfidencePatterns := []string{
		"fetch\\(", "axios\\.", "\\.get\\(", "\\.post\\(",
		"app\\.(get|post)", "router\\.(get|post)",
	}
	
	for _, highPattern := range highConfidencePatterns {
		if strings.Contains(pattern, highPattern) {
			return 0.9
		}
	}
	
	// Medium confidence patterns
	mediumConfidencePatterns := []string{
		"/api/", "/admin/", "/auth/",
	}
	
	for _, mediumPattern := range mediumConfidencePatterns {
		if strings.Contains(pattern, mediumPattern) {
			return 0.7
		}
	}
	
	// Default confidence
	return 0.5
}