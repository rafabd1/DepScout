package core

import "regexp"

/**
 * @description Pre-compiled regex patterns optimized for Go's regex engine
 * These patterns are designed to avoid catastrophic backtracking and handle
 * various JavaScript syntax nuances efficiently
 */

/**
 * @description Compiles endpoint extraction patterns optimized for performance
 * @returns Array of compiled regex patterns for endpoint detection
 */
func compileEndpointPatterns() []*regexp.Regexp {
	patterns := []string{
		// Standard REST API endpoints with path parameters
		`["'\x60](/(?:[a-zA-Z0-9_\-]+/)*[a-zA-Z0-9_\-]*(?:\{[a-zA-Z0-9_]+\})?(?:/[a-zA-Z0-9_\-]*)*)["\x60']`,
		
		// Template literal endpoints with variables
		`\x60(/[^/\x60]*(?:\$\{[^}]+\}[^/\x60]*)*/?[^/\x60]*)\x60`,
		
		// Fetch/axios URL patterns
		`(?:fetch|axios(?:\.get|\.post|\.put|\.delete)?)\s*\(\s*["'\x60]([/][^"'\x60]*)["\x60']`,
		
		// XMLHttpRequest open calls
		`\.open\s*\(\s*["'][^"']*["']\s*,\s*["'\x60]([/][^"'\x60]*)["\x60']`,
		
		// API versioning patterns
		`["'\x60](/(?:api|v\d+)(?:/[a-zA-Z0-9_\-]+)*)["\x60']`,
		
		// Route definitions (Express.js style)
		`\.(?:get|post|put|delete|patch)\s*\(\s*["'\x60]([/][^"'\x60]*)["\x60']`,
		
		// URL constructor patterns
		`new\s+URL\s*\(\s*["'\x60]([/][^"'\x60]*)["\x60']`,
		
		// Relative path patterns
		`["'\x60](\./[a-zA-Z0-9_/\-\.]+)["\x60']`,
		
		// Common endpoint patterns in strings
		`["'\x60](/admin[/a-zA-Z0-9_\-]*)["\x60']`,
		`["'\x60](/api[/a-zA-Z0-9_\-]*)["\x60']`,
		`["'\x60](/auth[/a-zA-Z0-9_\-]*)["\x60']`,
		`["'\x60](/user[s]?[/a-zA-Z0-9_\-]*)["\x60']`,
		
		// GraphQL endpoints
		`["'\x60]([/]graphql[/]?)["\x60']`,
		
		// WebSocket endpoints
		`["'\x60]([/]ws[/a-zA-Z0-9_\-]*)["\x60']`,
	}

	return compilePatterns(patterns)
}

/**
 * @description Compiles parameter extraction patterns for various contexts
 * @returns Array of compiled regex patterns for parameter detection
 */
func compileParameterPatterns() []*regexp.Regexp {
	patterns := []string{
		// URL query parameters
		`[?&]([a-zA-Z0-9_]+)=`,
		
		// Object property parameters (JSON-like)
		`["'\x60]([a-zA-Z0-9_]+)["'\x60]\s*:\s*(?:["'\x60][^"'\x60]*["'\x60]|[^,}]+)`,
		
		// Function parameter destructuring
		`\{\s*([a-zA-Z0-9_]+)(?:\s*,\s*([a-zA-Z0-9_]+))*\s*\}`,
		
		// Form field names
		`name\s*=\s*["'\x60]([a-zA-Z0-9_]+)["'\x60']`,
		
		// Input field IDs
		`id\s*=\s*["'\x60]([a-zA-Z0-9_]+)["'\x60']`,
		
		// Data attributes
		`data-([a-zA-Z0-9_\-]+)`,
		
		// Parameter in template literals
		`\$\{([a-zA-Z0-9_]+)\}`,
		
		// Common parameter patterns
		`\.get\s*\(\s*["'\x60]([a-zA-Z0-9_]+)["'\x60']`,
		
		// POST body parameters (simple detection)
		`([a-zA-Z0-9_]+)\s*:\s*[a-zA-Z0-9_$.]+`,
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