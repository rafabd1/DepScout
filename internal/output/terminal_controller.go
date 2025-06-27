package output

import (
	"fmt"
	"os"
	"sync"

	"DepScout/internal/utils"
)

// TerminalController gerencia o acesso exclusivo à saída do terminal,
// coordenando entre a barra de progresso e as mensagens de log.
type TerminalController struct {
	mu              sync.Mutex
	outputMu        sync.Mutex // Garante que apenas uma coisa (log ou barra) escreva por vez.
	isTerminal      bool
	activeProgressBar *ProgressBar // Referência direta para a barra de progresso ativa.
}

var (
	instance *TerminalController
	once     sync.Once
)

// GetTerminalController retorna a instância singleton do TerminalController.
func GetTerminalController() *TerminalController {
	once.Do(func() {
		instance = &TerminalController{
			isTerminal: utils.IsTerminal(os.Stderr.Fd()),
		}
	})
	return instance
}

// BeginOutput bloqueia a saída para uso exclusivo.
func (tc *TerminalController) BeginOutput() {
	tc.outputMu.Lock()
}

// EndOutput libera o bloqueio de saída.
func (tc *TerminalController) EndOutput() {
	tc.outputMu.Unlock()
}

// SetActiveProgressBar define ou limpa a barra de progresso ativa.
func (tc *TerminalController) SetActiveProgressBar(pb *ProgressBar) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.activeProgressBar = pb
}

// GetActiveProgressBar retorna a barra de progresso atualmente ativa como uma interface.
func (tc *TerminalController) GetActiveProgressBar() utils.ProgressBarInterface {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.activeProgressBar == nil {
		return nil
	}
	return tc.activeProgressBar
}

// ClearLine limpa a linha atual do terminal, se for um terminal.
func (tc *TerminalController) ClearLine() {
	if tc.isTerminal {
		fmt.Fprint(os.Stderr, "\033[2K\r")
	}
}

// CoordinateOutput executa uma função com acesso exclusivo à saída do terminal.
func (tc *TerminalController) CoordinateOutput(fn func()) {
	tc.BeginOutput()
	defer tc.EndOutput()

	// O chamador (logger) agora é responsável por pausar/retomar a barra.
	// Esta função apenas garante que a saída seja serializada.
	fn()
}

// IsTerminal retorna se o controlador está anexado a um terminal.
func (tc *TerminalController) IsTerminal() bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.isTerminal
} 