package core

import (
	"strings"

	"github.com/dop251/goja/ast"
	"github.com/dop251/goja/parser"
)

/**
 * @description AST Engine for JavaScript analysis - simplified and functional
 */
type ASTEngine struct{}

/**
 * @description Creates a new AST Engine
 */
func NewASTEngine() *ASTEngine {
	return &ASTEngine{}
}

/**
 * @description Enhances regex findings with AST analysis
 */
func (ae *ASTEngine) EnhanceFindings(content []byte, regexFindings []RawFinding) []RawFinding {
	// Try AST parsing
	program, err := parser.ParseFile(nil, "", string(content), 0)
	if err != nil {
		// Return regex findings with fallback context
		for i := range regexFindings {
			regexFindings[i].Context = "regex_only"
		}
		return regexFindings
	}

	// Extract AST findings
	astFindings := ae.extractFromAST(program)
	
	// Merge findings
	return ae.mergeFindings(astFindings, regexFindings)
}

/**
 * @description Extract findings from AST
 */
func (ae *ASTEngine) extractFromAST(program *ast.Program) []RawFinding {
	collector := &findingCollector{
		endpoints: make([]Endpoint, 0),
		domains:   make([]string, 0),
	}

	ae.walkProgram(program, collector)

	if len(collector.endpoints) > 0 || len(collector.domains) > 0 {
		return []RawFinding{{
			Endpoints:  collector.endpoints,
			Domains:    collector.domains,
			Context:    "ast_analysis",
			Confidence: 0.9,
		}}
	}

	return []RawFinding{}
}

/**
 * @description Walk AST program
 */
func (ae *ASTEngine) walkProgram(program *ast.Program, collector *findingCollector) {
	for _, stmt := range program.Body {
		ae.walkNode(stmt, collector)
	}
}

/**
 * @description Walk AST nodes recursively
 */
func (ae *ASTEngine) walkNode(node ast.Node, collector *findingCollector) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.ExpressionStatement:
		ae.walkNode(n.Expression, collector)
	case *ast.CallExpression:
		ae.handleCallExpression(n, collector)
		ae.walkNode(n.Callee, collector)
		for _, arg := range n.ArgumentList {
			ae.walkNode(arg, collector)
		}
	case *ast.BlockStatement:
		for _, stmt := range n.List {
			ae.walkNode(stmt, collector)
		}
	case *ast.FunctionDeclaration:
		if n.Function != nil {
			ae.walkNode(n.Function.Body, collector)
		}
	case *ast.VariableStatement:
		for _, decl := range n.List {
			if decl.Initializer != nil {
				ae.walkNode(decl.Initializer, collector)
			}
		}
	}
}

/**
 * @description Handle function calls
 */
func (ae *ASTEngine) handleCallExpression(call *ast.CallExpression, collector *findingCollector) {
	if !ae.isHTTPCall(call) {
		return
	}

	if len(call.ArgumentList) == 0 {
		return
	}

	// Extract URL from first argument
	urlArg := ae.getStringValue(call.ArgumentList[0])
	if urlArg == "" || !ae.looksLikeEndpoint(urlArg) {
		return
	}

	endpoint := Endpoint{
		Path:    urlArg,
		Method:  ae.inferMethod(call),
		Context: "ast_call",
	}
	collector.endpoints = append(collector.endpoints, endpoint)

	// Extract domain if present
	if domain := ae.extractDomainFromURL(urlArg); domain != "" {
		collector.domains = append(collector.domains, domain)
	}
}

/**
 * @description Check if call is HTTP-related
 */
func (ae *ASTEngine) isHTTPCall(call *ast.CallExpression) bool {
	if ident, ok := call.Callee.(*ast.Identifier); ok {
		name := ident.Name.String()
		return name == "fetch" || name == "axios" || name == "request"
	}

	if member, ok := call.Callee.(*ast.DotExpression); ok {
		methodName := member.Identifier.Name.String()
		return methodName == "get" || methodName == "post" || 
		       methodName == "put" || methodName == "delete" || 
		       methodName == "patch" || methodName == "ajax"
	}

	return false
}

/**
 * @description Infer HTTP method from call
 */
func (ae *ASTEngine) inferMethod(call *ast.CallExpression) string {
	if member, ok := call.Callee.(*ast.DotExpression); ok {
		method := strings.ToUpper(member.Identifier.Name.String())
		if method == "GET" || method == "POST" || method == "PUT" || 
		   method == "DELETE" || method == "PATCH" {
			return method
		}
	}
	return "GET"
}

/**
 * @description Get string value from AST node
 */
func (ae *ASTEngine) getStringValue(node ast.Node) string {
	if node == nil {
		return ""
	}

	switch n := node.(type) {
	case *ast.StringLiteral:
		return n.Value.String()
	case *ast.Identifier:
		return n.Name.String()
	}

	return ""
}

/**
 * @description Check if string looks like endpoint
 */
func (ae *ASTEngine) looksLikeEndpoint(path string) bool {
	return len(path) > 1 && (strings.HasPrefix(path, "/") || strings.Contains(path, "://"))
}

/**
 * @description Extract domain from URL
 */
func (ae *ASTEngine) extractDomainFromURL(url string) string {
	if !strings.Contains(url, "://") {
		return ""
	}

	parts := strings.Split(url, "://")
	if len(parts) < 2 {
		return ""
	}

	domainPart := strings.Split(parts[1], "/")[0]
	return domainPart
}

/**
 * @description Merge AST and regex findings
 */
func (ae *ASTEngine) mergeFindings(astFindings, regexFindings []RawFinding) []RawFinding {
	if len(astFindings) == 0 {
		return regexFindings
	}

	if len(regexFindings) == 0 {
		return astFindings
	}

	// Combine findings
	enhanced := make([]RawFinding, 0, len(regexFindings))
	astFinding := astFindings[0]

	for _, regexFinding := range regexFindings {
		combined := regexFinding
		combined.Context = "hybrid_analysis"
		combined.Confidence = 0.85

		// Merge endpoints with method enhancement
		combined.Endpoints = ae.mergeEndpoints(regexFinding.Endpoints, astFinding.Endpoints)
		
		// Merge domains
		combined.Domains = ae.mergeDomains(regexFinding.Domains, astFinding.Domains)

		enhanced = append(enhanced, combined)
	}

	return enhanced
}

/**
 * @description Merge endpoints from different sources
 */
func (ae *ASTEngine) mergeEndpoints(regexEps, astEps []Endpoint) []Endpoint {
	endpointMap := make(map[string]Endpoint)

	// Add regex endpoints
	for _, ep := range regexEps {
		endpointMap[ep.Path] = ep
	}

	// Enhance with AST endpoints
	for _, ep := range astEps {
		if existing, exists := endpointMap[ep.Path]; exists {
			// Prefer AST method if more specific
			if ep.Method != "GET" || existing.Method == "GET" {
				existing.Method = ep.Method
			}
			existing.Context = "hybrid"
			endpointMap[ep.Path] = existing
		} else {
			endpointMap[ep.Path] = ep
		}
	}

	// Convert back to slice
	result := make([]Endpoint, 0, len(endpointMap))
	for _, ep := range endpointMap {
		result = append(result, ep)
	}

	return result
}

/**
 * @description Merge domains from different sources
 */
func (ae *ASTEngine) mergeDomains(regexDomains, astDomains []string) []string {
	domainSet := make(map[string]bool)
	
	for _, domain := range regexDomains {
		domainSet[domain] = true
	}
	
	for _, domain := range astDomains {
		domainSet[domain] = true
	}

	result := make([]string, 0, len(domainSet))
	for domain := range domainSet {
		result = append(result, domain)
	}

	return result
}

/**
 * @description Collector for findings during AST traversal
 */
type findingCollector struct {
	endpoints []Endpoint
	domains   []string
}