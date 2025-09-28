package core

import "regexp"

/**
 * @description Pre-compiled regex patterns optimized for Go's regex engine
 * These patterns are designed to avoid catastrophic backtracking and handle
 * various JavaScript syntax nuances efficiently
 */

/**
 * @description Compiles endpoint extraction patterns optimized for performance
 * Patterns are highly specific to avoid JavaScript code false positives
 * @returns Array of compiled regex patterns for endpoint detection
 */
func compileEndpointPatterns() []*regexp.Regexp {
	patterns := []string{
		// Fetch/axios URL patterns - strict context
		`(?:fetch|axios(?:\.(?:get|post|put|delete|patch))?)\s*\(\s*["'\x60]([/](?:[a-zA-Z0-9_\-]+[/]?)*(?:\{[a-zA-Z0-9_]+\})?[/]?[a-zA-Z0-9_\-]*)["\x60']`,
		
		// XMLHttpRequest open calls - complete pattern
		`\.open\s*\(\s*["'](?:GET|POST|PUT|DELETE|PATCH)["']\s*,\s*["'\x60]([/][a-zA-Z0-9_/\-\.]+)["\x60']`,
		
		// URL constructor with API paths
		`new\s+URL\s*\(\s*["'\x60]([/](?:api|auth|admin|user)[a-zA-Z0-9_/\-]*)["\x60']`,
		
		// Route definitions (Express.js/server-side)
		`\.(?:get|post|put|delete|patch)\s*\(\s*["'\x60]([/][a-zA-Z0-9_/\-:{}]+)["\x60']\s*,`,
		
		// API versioning patterns - specific
		`["'\x60]([/](?:api|v[1-9])[/][a-zA-Z0-9_/\-]+)["\x60']`,
		
		// RESTful endpoints - common patterns
		`["'\x60]([/](?:api|admin|auth|users?|posts?|products?|orders?)[/]?[a-zA-Z0-9_/\-]*)["\x60']`,
		
		// GraphQL endpoints
		`["'\x60]([/]graphql(?:[/]v[1-9])?[/]?)["\x60']`,
		
		// WebSocket endpoints
		`["'\x60]([/](?:ws|websocket|socket)[/]?[a-zA-Z0-9_\-]*)["\x60']`,
		
		// Action/form endpoints
		`action\s*=\s*["'\x60]([/][a-zA-Z0-9_/\-\.]+)["\x60']`,
		
		// AJAX URL patterns with clear context
		`\.(?:ajax|load|submit)\s*\([^)]*url\s*:\s*["'\x60]([/][a-zA-Z0-9_/\-\.]+)["\x60']`,
		
		// Configuration URL patterns
		`(?:baseURL|endpoint|apiUrl)\s*[:=]\s*["'\x60]([/][a-zA-Z0-9_/\-\.]+)["\x60']`,
		
		// Template literals with simple variables (not complex expressions)
		`\x60([/][a-zA-Z0-9_/\-]*\$\{[a-zA-Z0-9_]+\}[a-zA-Z0-9_/\-]*)\x60`,
	}

	return compilePatterns(patterns)
}

/**
 * @description Compiles parameter extraction patterns for various contexts
 * Patterns are highly specific to avoid false positives from JavaScript code
 * @returns Array of compiled regex patterns for parameter detection
 */
func compileParameterPatterns() []*regexp.Regexp {
	patterns := []string{
		// 1. Explicit URL query parameters in strings (very specific)
		`["'\x60][^"'\x60]*\?[^"'\x60]*[&]([a-zA-Z][a-zA-Z0-9_\-]{1,20})=[\w\-\.%]*["'\x60]`,
		`["'\x60][^"'\x60]*\?([a-zA-Z][a-zA-Z0-9_\-]{1,20})=[\w\-\.%]*["'\x60]`,
		
		// 2. URLSearchParams constructor - explicit API context
		`new\s+URLSearchParams\s*\(\s*\{[^}]*["'\x60]([a-zA-Z][a-zA-Z0-9_\-]{1,20})["'\x60']\s*:\s*[^}]*\}`,
		
		// 3. FormData append - very specific API calls
		`\.append\s*\(\s*["'\x60]([a-zA-Z][a-zA-Z0-9_\-]{1,20})["'\x60']\s*,`,
		
		// 4. Fetch/axios params - in HTTP request context
		`(?:fetch|axios)\s*\([^,]*,\s*\{[^}]*(?:params|data)\s*:\s*\{[^}]*["'\x60]([a-zA-Z][a-zA-Z0-9_\-]{1,20})["'\x60']\s*:`,
		
		// 5. HTML form inputs only
		`<input[^>]+name\s*=\s*["'\x60]([a-zA-Z][a-zA-Z0-9_\-]{1,20})["'\x60']`,
		
		// 6. REST API path parameters - in URL context
		`["'\x60][^"'\x60]*\/\{([a-zA-Z][a-zA-Z0-9_]{1,15})\}[^"'\x60]*["'\x60]`,
		
		// 7. GraphQL variables - $ prefix in operation
		`\$([a-zA-Z][a-zA-Z0-9_]{1,20}):\s*[A-Z]\w+[!\]?`,
		
		// 8. Query parameters in actual URL strings
		`(?:http|\/api|\/v\d)[^"'\x60]*[?&]([a-zA-Z][a-zA-Z0-9_\-]{1,20})=`,
	}

	return compilePatterns(patterns)
}

/**
 * @description Compiles header extraction patterns for HTTP headers
 * @returns Array of compiled regex patterns for header detection
 */
func compileHeaderPatterns() []*regexp.Regexp {
	patterns := []string{
		// Standard header setting patterns
		`\.setRequestHeader\s*\(\s*["'\x60]([A-Za-z\-]+)["'\x60']\s*,\s*["'\x60]([^"'\x60]*)["'\x60']`,
		
		// Headers object properties
		`headers\s*:\s*\{[^}]*["'\x60]([A-Za-z\-]+)["'\x60']\s*:\s*["'\x60]([^"'\x60]*)["'\x60']`,
		
		// Authorization headers
		`["'\x60](Authorization)["'\x60']\s*:\s*["'\x60]([^"'\x60]*)["'\x60']`,
		
		// Content-Type headers
		`["'\x60](Content-Type)["'\x60']\s*:\s*["'\x60]([^"'\x60]*)["'\x60']`,
		
		// Bearer token patterns
		`["'\x60](Authorization)["'\x60']\s*:\s*["'\x60](Bearer\s+[^"'\x60]*)["'\x60']`,
		
		// Custom headers (X- prefix)
		`["'\x60](X-[A-Za-z\-]+)["'\x60']\s*:\s*["'\x60]([^"'\x60]*)["'\x60']`,
		
		// API key headers
		`["'\x60]([A-Za-z\-]*[Aa]pi[Kk]ey[A-Za-z\-]*)["'\x60']\s*:\s*["'\x60]([^"'\x60]*)["'\x60']`,
		
		// Accept headers
		`["'\x60](Accept)["'\x60']\s*:\s*["'\x60]([^"'\x60]*)["'\x60']`,
		
		// User-Agent headers
		`["'\x60](User-Agent)["'\x60']\s*:\s*["'\x60]([^"'\x60]*)["'\x60']`,
	}

	return compilePatterns(patterns)
}

/**
 * @description Compiles domain extraction patterns for various URL formats
 * @returns Array of compiled regex patterns for domain detection
 */
func compileDomainPatterns() []*regexp.Regexp {
	patterns := []string{
		// Full URLs with protocol
		`https?://([a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(?:\.[a-zA-Z]{2,})?)`,
		
		// Domain-only patterns
		`["'\x60]([a-zA-Z0-9\-]+\.(?:com|net|org|edu|gov|mil|int|co\.uk|io|dev|app))["\x60']`,
		
		// Subdomain patterns
		`["'\x60]([a-zA-Z0-9\-]+\.(?:api|admin|www|app|dev|test|staging)\.[a-zA-Z0-9\-\.]+)["\x60']`,
		
		// Internal domain patterns
		`["'\x60]([a-zA-Z0-9\-]+\.(?:local|internal|corp|intranet))["\x60']`,
		
		// API endpoint domains
		`["'\x60](api\.[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,})["\x60']`,
		
		// CDN patterns
		`["'\x60]([a-zA-Z0-9\-]+\.(?:cdn|assets)\.[a-zA-Z0-9\-\.]+)["\x60']`,
	}

	return compilePatterns(patterns)
}

/**
 * @description Compiles HTTP method detection patterns
 * @returns Array of compiled regex patterns for method detection
 */
func compileMethodPatterns() []*regexp.Regexp {
	patterns := []string{
		// Explicit method declarations
		`method\s*:\s*["'\x60](GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)["'\x60']`,
		
		// HTTP method words in context
		`\b(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\b`,
		
		// Axios method calls
		`axios\.(get|post|put|delete|patch)`,
		
		// Fetch method options
		`method\s*:\s*["'\x60]([A-Z]+)["'\x60']`,
		
		// XMLHttpRequest method
		`\.open\s*\(\s*["'\x60]([A-Z]+)["'\x60']`,
	}

	return compilePatterns(patterns)
}

/**
 * @description Compiles request body extraction patterns
 * @returns Array of compiled regex patterns for body detection
 */
func compileBodyPatterns() []*regexp.Regexp {
	patterns := []string{
		// JSON body patterns
		`body\s*:\s*JSON\.stringify\s*\(([^)]+)\)`,
		
		// Direct object body
		`body\s*:\s*(\{[^}]+\})`,
		
		// FormData patterns
		`body\s*:\s*(new\s+FormData\([^)]*\))`,
		
		// URLSearchParams
		`body\s*:\s*(new\s+URLSearchParams\([^)]*\))`,
		
		// String body
		`body\s*:\s*["'\x60]([^"'\x60]+)["'\x60']`,
		
		// Send method body
		`\.send\s*\(([^)]+)\)`,
	}

	return compilePatterns(patterns)
}

/**
 * @description Compiles HTML-specific patterns for script tags, forms, etc.
 * @returns Array of compiled regex patterns for HTML element detection
 */
func compileHTMLPatterns() []*regexp.Regexp {
	patterns := []string{
		// Script src attributes
		`<script[^>]*src\s*=\s*["'\x60]([^"'\x60]+)["'\x60']`,
		
		// Link href attributes
		`<link[^>]*href\s*=\s*["'\x60]([^"'\x60]+)["'\x60']`,
		
		// Form action attributes
		`<form[^>]*action\s*=\s*["'\x60]([^"'\x60]+)["'\x60']`,
		
		// Iframe src attributes
		`<iframe[^>]*src\s*=\s*["'\x60]([^"'\x60]+)["'\x60']`,
		
		// Image src attributes
		`<img[^>]*src\s*=\s*["'\x60]([^"'\x60]+)["'\x60']`,
		
		// AJAX calls in script tags
		`\$\.(?:get|post|ajax)\s*\(\s*["'\x60]([^"'\x60]+)["'\x60']`,
		
		// Data attributes with URLs
		`data-[a-zA-Z\-]*url\s*=\s*["'\x60]([^"'\x60]+)["'\x60']`,
	}

	return compilePatterns(patterns)
}

/**
 * @description Compiles JSON-specific patterns for configuration files
 * @returns Array of compiled regex patterns for JSON structure detection
 */
func compileJSONPatterns() []*regexp.Regexp {
	patterns := []string{
		// URL properties in JSON
		`["'\x60](?:url|endpoint|path)["'\x60']\s*:\s*["'\x60]([^"'\x60]+)["'\x60']`,
		
		// API configuration
		`["'\x60](?:api|base|root)["'\x60']\s*:\s*["'\x60]([^"'\x60]+)["'\x60']`,
		
		// Server configuration
		`["'\x60](?:server|host)["'\x60']\s*:\s*["'\x60]([^"'\x60]+)["'\x60']`,
		
		// Proxy configuration
		`["'\x60]proxy["'\x60']\s*:\s*["'\x60]([^"'\x60]+)["'\x60']`,
		
		// Routes in configuration
		`["'\x60]routes?["'\x60']\s*:\s*\{[^}]*["'\x60]([^"'\x60/]+)["'\x60']`,
		
		// Paths in package.json scripts
		`["'\x60](?:start|build|dev)["'\x60']\s*:\s*["'\x60]([^"'\x60]+)["'\x60']`,
	}

	return compilePatterns(patterns)
}

/**
 * @description Helper function to compile pattern strings into regexp objects
 * @param patterns Array of pattern strings
 * @returns Array of compiled regex patterns
 */
func compilePatterns(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	
	for _, pattern := range patterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			compiled = append(compiled, regex)
		}
		// Note: In production, you might want to log compilation errors
		// For now, we silently skip invalid patterns
	}
	
	return compiled
}