package output

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type updateOp int

const (
	opIncrement updateOp = iota
	opIncrementTotal
	opUpdate
	opReset
)

type progressUpdate struct {
	op    updateOp
	value int
}

// ProgressBar é uma barra de progresso para o terminal.
type ProgressBar struct {
	total            int
	current          int
	width            int
	refresh          time.Duration
	startTime        time.Time
	mu               sync.Mutex
	done             chan struct{}
	writer           io.Writer
	autoRefresh      bool
	isActive         bool
	spinner          int
	spinnerChars     []string
	prefix           string
	suffix           string
	isTerminal       bool
	renderPaused     bool
	outputControl    chan struct{}
	terminalController *TerminalController // Referência ao controlador
	updateChan       chan progressUpdate
}

// NewProgressBar cria uma nova ProgressBar.
func NewProgressBar(total int, width int) *ProgressBar {
	tc := GetTerminalController()

	return &ProgressBar{
		total:              total,
		current:            0,
		width:              width,
		refresh:            150 * time.Millisecond,
		done:               make(chan struct{}),
		writer:             os.Stderr,
		isActive:           false,
		spinnerChars:       []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		prefix:             "",
		suffix:             "",
		isTerminal:         tc.IsTerminal(),
		renderPaused:       false,
		outputControl:      make(chan struct{}, 1),
		terminalController: tc,
		updateChan:         make(chan progressUpdate, 100), // Buffer para não bloquear workers
	}
}

// Start inicia a renderização da barra de progresso.
func (pb *ProgressBar) Start() {
	pb.mu.Lock()
	if pb.isActive {
		pb.mu.Unlock()
		return
	}

	pb.startTime = time.Now()
	pb.isActive = true
	pb.mu.Unlock()

	// Registra esta instância da barra como ativa no controlador do terminal.
	pb.terminalController.SetActiveProgressBar(pb)

	if pb.isTerminal {
		go pb.renderLoop()
	}
}

// Stop para a barra de progresso e a limpa do terminal.
func (pb *ProgressBar) Stop() {
	pb.mu.Lock()
	if !pb.isActive {
		pb.mu.Unlock()
		return
	}
	pb.isActive = false

	select {
	case <-pb.done:
		// já fechado
	default:
		close(pb.done)
	}

	// Desregistra a barra do controlador de terminal
	pb.terminalController.SetActiveProgressBar(nil)
	pb.mu.Unlock()

	// Pequena espera para garantir que as goroutines de renderização parem.
	time.Sleep(50 * time.Millisecond)

	if pb.isTerminal {
		pb.terminalController.BeginOutput()
		pb.clearBar()
		pb.terminalController.EndOutput()
	}
}

// Finalize é um wrapper para Stop, para manter a compatibilidade da API.
func (pb *ProgressBar) Finalize() {
	pb.Stop()
}

func (pb *ProgressBar) renderLoop() {
	ticker := time.NewTicker(pb.refresh)
	defer ticker.Stop()

	for {
		select {
		case <-pb.done:
			close(pb.updateChan)
			// Drena quaisquer atualizações restantes
			for range pb.updateChan {
			}
			return
		case update := <-pb.updateChan:
			switch update.op {
			case opIncrement:
				pb.current++
			case opIncrementTotal:
				pb.total += update.value
			case opUpdate:
				pb.current = update.value
			case opReset:
				pb.total = update.value
				pb.current = 0
				pb.startTime = time.Now()
			}
		case <-ticker.C:
			pb.actualRender()
		}
	}
}

func (pb *ProgressBar) actualRender() {
	pb.terminalController.BeginOutput()
	defer pb.terminalController.EndOutput()

	pb.mu.Lock()
	
	// Verificações redundantes, mas seguras
	if !pb.isActive || !pb.isTerminal || pb.renderPaused {
		pb.mu.Unlock()
		return
	}
	
	pb.spinner = (pb.spinner + 1) % len(pb.spinnerChars)
	
	currentTotal := pb.total
	currentProgress := pb.current
	if currentTotal == 0 { // Evita divisão por zero se total for 0 (ex: nenhum job)
		currentProgress = 0 // Garante 0% se total for 0
	}

	percent := 0.0
	if currentTotal > 0 {
		percent = float64(currentProgress) / float64(currentTotal) * 100
	}
	
	elapsed := time.Since(pb.startTime)
	
	var etaStr string
	if currentProgress > 0 && currentProgress < currentTotal {
		eta := time.Duration(float64(elapsed) * float64(currentTotal-currentProgress) / float64(currentProgress))
		etaStr = formatDuration(eta)
	} else if currentProgress >= currentTotal && currentTotal > 0 { // Concluído
		etaStr = "Done"
	} else { // currentProgress == 0 ou total == 0
		etaStr = "N/A"
	}
	
	completedWidth := 0
	if currentTotal > 0 {
		completedWidth = int(float64(pb.width) * float64(currentProgress) / float64(currentTotal))
	}
	if completedWidth > pb.width {
		completedWidth = pb.width
	}
	if completedWidth < 0 {
		completedWidth = 0
	}
	
	bar := strings.Repeat("█", completedWidth) + strings.Repeat("░", pb.width-completedWidth)
	
	status := fmt.Sprintf("%s%s [%s] %d/%d (%.2f%%) | Elapsed: %s | ETA: %s %s",
		pb.prefix,
		pb.spinnerChars[pb.spinner],
		bar,
		currentProgress, currentTotal,
		percent,
		formatDuration(elapsed),
		etaStr,
		pb.suffix,
	)
	
	// Não precisamos de lastPrintedChars se sempre usamos \033[2K\r
	// pb.lastPrintedChars = len(status) 
	pb.mu.Unlock()
	
	// DEBUG: Temporarily log to understand flickering
	// fmt.Fprintf(os.Stderr, "[DEBUG PB RENDER] Instance: %p, Prefix: '%s', Current: %d, Total: %d, IsActive: %t, Paused: %t, Spinner: %s\n", pb, pb.prefix, currentProgress, currentTotal, pb.isActive, pb.renderPaused, pb.spinnerChars[pb.spinner])

	fmt.Fprint(pb.writer, "\033[2K\r"+status) // Limpa linha, volta ao início, imprime status
}

// MoveForLog pausa e limpa a barra para permitir a escrita de um log.
func (pb *ProgressBar) MoveForLog() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if !pb.isActive || !pb.isTerminal {
		return
	}
	pb.renderPaused = true
	pb.clearBar()
}

// ShowAfterLog redesenha a barra após um log ter sido escrito.
func (pb *ProgressBar) ShowAfterLog() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if !pb.isActive || !pb.isTerminal {
		return
	}
	pb.renderPaused = false
	pb.actualRender()
}

func (pb *ProgressBar) clearBar() {
	fmt.Fprint(pb.writer, "\033[2K\r")
}

func (pb *ProgressBar) IsTerminal() bool {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	return pb.isTerminal
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	s := d.Seconds()
	if s < 0 { s = 0 } // Evita durações negativas na exibição

	if s < 60 {
		return fmt.Sprintf("%.0fs", s)
	}
	
	m := int(s/60) % 60
	h := int(s/3600)
	sRemaining := int(s) % 60

	if h < 1 {
		return fmt.Sprintf("%dm%02ds", m, sRemaining)
	}
	
	return fmt.Sprintf("%dh%02dm%02ds", h, m, sRemaining)
}

func (pb *ProgressBar) GetPrefixForDebug() string {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	return pb.prefix
}

// Increment avança a barra de progresso em um.
func (pb *ProgressBar) Increment() {
	if pb.isTerminal {
		pb.updateChan <- progressUpdate{op: opIncrement}
	}
}

// IncrementTotal aumenta o total da barra de progresso.
func (pb *ProgressBar) IncrementTotal(n int) {
	if pb.isTerminal {
		pb.updateChan <- progressUpdate{op: opIncrementTotal, value: n}
	}
}

// Update define o valor atual da barra de progresso.
func (pb *ProgressBar) Update(current int) {
	if pb.isTerminal {
		pb.updateChan <- progressUpdate{op: opUpdate, value: current}
	}
}

// SetTotalAndReset permite redefinir o total da barra e zerar o progresso.
func (pb *ProgressBar) SetTotalAndReset(newTotal int) {
	if pb.isTerminal {
		pb.updateChan <- progressUpdate{op: opReset, value: newTotal}
	}
}

// PauseRender impede temporariamente que a barra seja redesenhada.
func (pb *ProgressBar) PauseRender() {
	pb.mu.Lock()
	pb.renderPaused = true
	// Limpa a barra ao pausar para não deixar uma barra estática enquanto logs são impressos.
	// Isso será feito por MoveForLog.
	pb.mu.Unlock()
}

// ResumeRender permite que a barra seja redesenhada novamente.
func (pb *ProgressBar) ResumeRender() {
	pb.mu.Lock()
	wasRenderPaused := pb.renderPaused
	pb.renderPaused = false
	pb.mu.Unlock()
	
	if wasRenderPaused && pb.isTerminal { // Só renderiza se estava pausado e é terminal
		pb.actualRender()
	}
} 