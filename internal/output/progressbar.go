package output

import (
	"fmt"
	"sync"
	"time"

	"github.com/morikuni/aec"
)

// ProgressBar displays a progress bar in the terminal.
type ProgressBar struct {
	controller *TerminalController
	total      int
	current    int
	mu         sync.Mutex
	ticker     *time.Ticker
	done       chan bool
}

// NewProgressBar creates a new ProgressBar.
func NewProgressBar(controller *TerminalController) *ProgressBar {
	return &ProgressBar{
		controller: controller,
		done:       make(chan bool),
	}
}

// Start begins rendering the progress bar.
func (p *ProgressBar) Start(total int) {
	p.total = total
	p.ticker = time.NewTicker(200 * time.Millisecond)
	go p.run()
}

// Stop halts the progress bar rendering.
func (p *ProgressBar) Stop() {
	p.ticker.Stop()
	p.done <- true
	// Clear the progress bar line
	p.controller.Printf("%s", aec.EraseLine(aec.EraseModes.All).String())
}

// Increment increases the progress count by one.
func (p *ProgressBar) Increment() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.current < p.total {
		p.current++
	}
}

func (p *ProgressBar) run() {
	p.controller.Println("") // Initial line for the progress bar to overwrite
	for {
		select {
		case <-p.done:
			return
		case <-p.ticker.C:
			p.render()
		}
	}
}

func (p *ProgressBar) render() {
	p.mu.Lock()
	defer p.mu.Unlock()

	percent := float64(p.current) / float64(p.total) * 100
	barWidth := 40
	filledWidth := int(float64(barWidth) * percent / 100)
	
	bar := ""
	for i := 0; i < filledWidth; i++ {
		bar += "█"
	}
	for i := filledWidth; i < barWidth; i++ {
		bar += " "
	}

	progressText := fmt.Sprintf("Progress: [%s] %d/%d (%.2f%%)", bar, p.current, p.total, percent)
	p.controller.Overwritef("%s", progressText)
} 