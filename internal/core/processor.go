package core

import (
	"regexp"
	"strings"
	"sync"

	"DepScout/internal/config"
	"DepScout/internal/utils"

	"github.com/dop251/goja/ast"
	"github.com/dop251/goja/parser"
)

// Processor is responsible for parsing JS files and finding package names.
type Processor struct {
	config          *config.Config
	logger          *utils.Logger
	scheduler       *Scheduler
	packageRegex    *regexp.Regexp
	checkedPackages sync.Map // Restaurado para evitar verificações duplicadas
}

// NewProcessor creates a new Processor instance.
func NewProcessor(cfg *config.Config, logger *utils.Logger) *Processor {
	// Regex aprimorada para validar o contexto da chamada `require` ou `import`.
	// Garante que "require" é seguido por parênteses e que "from" está em um contexto de importação.
	// Isso evita a captura de strings em outras funções ou construtos de linguagem.
	regex := regexp.MustCompile(`(?:require\s*\(\s*|import(?:["'a-zA-Z0-9\s{},*]+)\s*from\s+)['"\x60]((?:@[a-z0-9_.-]+\/)?[a-z0-9_.-]+)['"\x60]`)

	return &Processor{
		config:       cfg,
		logger:       logger,
		packageRegex: regex,
	}
}

// SetScheduler gives the processor a reference to the scheduler to add new jobs.
func (p *Processor) SetScheduler(s *Scheduler) {
	p.scheduler = s
}

// ProcessJSFileContent analisa o conteúdo de um arquivo JavaScript em busca de pacotes.
func (p *Processor) ProcessJSFileContent(sourceURL string, body []byte) error {
	p.logger.Debugf("Processing content from %s (%d bytes)", sourceURL, len(body))

	if p.config.DeepScan {
		return p.processWithAST(sourceURL, body)
	}
	return p.processWithRegex(sourceURL, body)
}

func (p *Processor) processWithRegex(sourceURL string, body []byte) error {
	matches := p.packageRegex.FindAllStringSubmatch(string(body), -1)
	if len(matches) == 0 {
		return nil
	}

	p.logger.Debugf("Found %d potential packages in %s", len(matches), sourceURL)

	for _, match := range matches {
		if len(match) > 1 {
			packageName := p.normalizePackageName(match[1])
			if p.isPackageWorthChecking(packageName) {
				// Apenas enfileira o job se o pacote não tiver sido verificado ainda.
				if _, loaded := p.checkedPackages.LoadOrStore(packageName, true); !loaded {
					// Corrigido: Cria o job manualmente para preservar a SourceURL original.
					verifyJob := Job{
						Input:     packageName,
						SourceURL: sourceURL,
						Type:      VerifyPackage,
					}
					p.scheduler.AddJobAsync(verifyJob)
				}
			}
		}
	}

	return nil
}

func (p *Processor) processWithAST(sourceURL string, body []byte) error {
	program, err := parser.ParseFile(nil, sourceURL, string(body), 0)
	if err != nil {
		p.logger.Warnf("Failed to parse AST for %s: %v. Falling back to regex.", sourceURL, err)
		return p.processWithRegex(sourceURL, body)
	}

	// Use a recursive function to traverse the AST
	p.traverseAST(program, sourceURL)
	return nil
}

// traverseAST recursively traverses the AST nodes looking for require() and import statements
func (p *Processor) traverseAST(node ast.Node, sourceURL string) {
	if node == nil {
		return
	}

	// Check for require() calls
	if call, ok := node.(*ast.CallExpression); ok {
		if ident, ok := call.Callee.(*ast.Identifier); ok && ident.Name == "require" {
			if len(call.ArgumentList) > 0 {
				if lit, ok := call.ArgumentList[0].(*ast.StringLiteral); ok {
					packageName := lit.Value.String()

					p.processFoundPackage(packageName, sourceURL)
				}

			}
		}
	}

	// Recursively traverse child nodes based on node type
	switch n := node.(type) {
	case *ast.Program:
		for _, stmt := range n.Body {
			p.traverseAST(stmt, sourceURL)
		}
	case *ast.BlockStatement:
		for _, stmt := range n.List {
			p.traverseAST(stmt, sourceURL)
		}
	case *ast.ExpressionStatement:
		p.traverseAST(n.Expression, sourceURL)
	case *ast.CallExpression:
		p.traverseAST(n.Callee, sourceURL)
		for _, arg := range n.ArgumentList {
			p.traverseAST(arg, sourceURL)
		}
	case *ast.FunctionDeclaration:
		if n.Function != nil {
			p.traverseAST(n.Function.Body, sourceURL)
		}
	case *ast.FunctionLiteral:
		p.traverseAST(n.Body, sourceURL)
	case *ast.LexicalDeclaration:
		for _, binding := range n.List {
			if binding.Initializer != nil {
				p.traverseAST(binding.Initializer, sourceURL)
			}
		}
	case *ast.VariableStatement:
		for _, decl := range n.List {
			p.traverseAST(decl, sourceURL)
		}
	case *ast.AssignExpression:
		p.traverseAST(n.Left, sourceURL)
		p.traverseAST(n.Right, sourceURL)
	case *ast.IfStatement:
		p.traverseAST(n.Test, sourceURL)
		p.traverseAST(n.Consequent, sourceURL)
		if n.Alternate != nil {
			p.traverseAST(n.Alternate, sourceURL)
		}
	case *ast.ForStatement:
		if n.Initializer != nil {
			p.traverseAST(n.Initializer, sourceURL)
		}
		if n.Test != nil {
			p.traverseAST(n.Test, sourceURL)
		}
		if n.Update != nil {
			p.traverseAST(n.Update, sourceURL)
		}
		p.traverseAST(n.Body, sourceURL)
	case *ast.ReturnStatement:
		if n.Argument != nil {
			p.traverseAST(n.Argument, sourceURL)
		}
	// Add more cases as needed for other node types
	}
}

// processFoundPackage handles a package name found in the AST
func (p *Processor) processFoundPackage(packageName, sourceURL string) {
	normalized := p.normalizePackageName(packageName)
	if p.isPackageWorthChecking(normalized) {
		if _, loaded := p.checkedPackages.LoadOrStore(normalized, true); !loaded {
			verifyJob := Job{
				Input:     normalized,
				SourceURL: sourceURL,
				Type:      VerifyPackage,
			}
			p.scheduler.AddJobAsync(verifyJob)
		}
	}

}

// normalizePackageName lida com casos como `pkg/subpath` -> `pkg`
// e preserva pacotes com escopo como `@scope/pkg`.
func (p *Processor) normalizePackageName(pkg string) string {
	pkg = strings.TrimSpace(pkg)
	if strings.HasPrefix(pkg, "@") {
		// Para pacotes com escopo como `@angular/core/testing`, queremos `@angular/core`.
		parts := strings.Split(pkg, "/")
		if len(parts) > 2 {
			return parts[0] + "/" + parts[1]
		}
	} else {
		// Para pacotes normais como `react-dom/client`, queremos `react-dom`.
		pkg = strings.Split(pkg, "/")[0]
	}
	return pkg
}

// isPackageWorthChecking filters out uninteresting package names.
func (p *Processor) isPackageWorthChecking(pkg string) bool {
	// Filtra pacotes relativos ou que não parecem ser do npm
	if strings.HasPrefix(pkg, ".") || strings.HasPrefix(pkg, "/") || strings.HasPrefix(pkg, "\\") {
		return false
	}
	// Filtra nomes de pacotes muito curtos
	if len(pkg) < 2 {
		return false
	}
	// Adicionar outros filtros se necessário (ex: extensões de arquivo)
	return true
} 
