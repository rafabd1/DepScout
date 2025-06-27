package core

import (
	"regexp"
	"strings"
	"sync"

	"DepScout/internal/config"
	"DepScout/internal/utils"
)

// Processor is responsible for parsing JS files and finding package names.
type Processor struct {
	config          *config.Config
	logger          utils.Logger
	scheduler       *Scheduler
	packageRegex    *regexp.Regexp
	checkedPackages sync.Map // Restaurado para evitar verificações duplicadas
}

// NewProcessor creates a new Processor instance.
func NewProcessor(cfg *config.Config, logger utils.Logger) *Processor {
	// Regex aprimorada para capturar dependências em `require` e `import`.
	// Cobre: require('pkg'), require("pkg"), require(`pkg`), import from 'pkg', etc.
	regex := regexp.MustCompile(`(?i)(?:require\s*\(\s*|import\s+.*?\s+from\s+)['"\x60]([^'"\x60]+)['"\x60]`)

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
					p.scheduler.addJob(NewJob(packageName, VerifyPackage))
				}
			}
		}
	}

	return nil
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