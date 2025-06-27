package output

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/morikuni/aec"
)

// ProgressBar displays a progress bar in the terminal.
type ProgressBar struct {
	controller        *TerminalController
	total             int
	current           int
	mu                *sync.Mutex
	ticker            *time.Ticker
	done              chan bool
	wg                sync.WaitGroup
	requestsPerSecond uint64 // Stored as bits of a float64
	startTime         time.Time
}

// NewProgressBar creates a new ProgressBar.
func NewProgressBar(controller *TerminalController) *ProgressBar {
    return &ProgressBar{
		controller: controller,
		done:       make(chan bool),
		startTime:  time.Now(), // Initialize start time
	}
}

// SetMutex sets the mutex to be used for synchronization.
func (p *ProgressBar) SetMutex(mu *sync.Mutex) {
	p.mu = mu
}

// SetRPS sets the current requests per second value.
func (p *ProgressBar) SetRPS(rps float64) {
	atomic.StoreUint64(&p.requestsPerSecond, math.Float64bits(rps))
                }

// Start begins rendering the progress bar.
func (p *ProgressBar) Start(total int) {
	p.total = total
	p.startTime = time.Now() // Reset start time when actually starting
	p.ticker = time.NewTicker(200 * time.Millisecond)
	p.wg.Add(1)
	go p.run()
}

// Stop halts the progress bar rendering.
func (p *ProgressBar) Stop() {
	if p.ticker == nil {
		return // Not started, nothing to stop.
	}
	// Stop the ticker and signal the run goroutine to finish.
	p.ticker.Stop()
	close(p.done)

	// Wait for the run goroutine to finish completely before clearing the line.
	p.wg.Wait()

	// Now that we are sure the run() goroutine is done, we can safely clear the line.
	p.Clear()
    }

// Increment increases the progress count by one.
func (p *ProgressBar) Increment() {
	// This operation is atomic on 64-bit systems for int, and for this use case,
	// a potential race on 32-bit systems is acceptable over a deadlock.
	// The mutex is for terminal I/O, not for this counter.
	if p.current < p.total {
		p.current++
	}
}

// Clear removes the progress bar from the terminal.
func (p *ProgressBar) Clear() {
	p.controller.Printf("\r%s", aec.EraseLine(aec.EraseModes.All))
}

// Render forces an immediate redraw of the progress bar.
// This is an internal method that assumes the caller holds the mutex.
func (p *ProgressBar) UnsafeRender() {
	p.render()
}

// run is the main loop for rendering the progress bar.
func (p *ProgressBar) run() {
	defer p.wg.Done()
	p.controller.Println("") // Initial line for the progress bar to overwrite
	for {
		select {
		case <-p.done:
			// Just before returning, render one last time to show 100%
			p.mu.Lock()
			p.render()
			p.mu.Unlock()
        return
		case <-p.ticker.C:
			p.mu.Lock()
			p.render()
			p.mu.Unlock()
		}
	}
}

// calculateETA estimates the remaining time based on current progress
func (p *ProgressBar) calculateETA() string {
	if p.current == 0 {
		return "calculating..."
	}
	
	elapsed := time.Since(p.startTime)
	if elapsed < time.Second {
		return "calculating..."
	}
	
	// Calculate progress rate (items per second)
	progressRate := float64(p.current) / elapsed.Seconds()
	if progressRate <= 0 {
		return "calculating..."
	}
	
	// Calculate remaining items and estimated time
	remaining := p.total - p.current
	if remaining <= 0 {
		return "done"
	}
	
	etaSeconds := float64(remaining) / progressRate
	etaDuration := time.Duration(etaSeconds * float64(time.Second))
	
	// Format ETA in a human-readable way
	if etaDuration < time.Minute {
		return fmt.Sprintf("%ds", int(etaDuration.Seconds()))
	} else if etaDuration < time.Hour {
		minutes := int(etaDuration.Minutes())
		seconds := int(etaDuration.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	} else {
		hours := int(etaDuration.Hours())
		minutes := int(etaDuration.Minutes()) % 60
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
}

// render draws the progress bar. It assumes the caller holds the mutex.
func (p *ProgressBar) render() {
	percent := float64(p.current) / float64(p.total) * 100
	if p.total == 0 { // Avoid division by zero
		percent = 100
	}
	barWidth := 40
	filledWidth := int(float64(barWidth) * percent / 100)

	bar := ""
	for i := 0; i < filledWidth; i++ {
		bar += "█"
	}
	for i := filledWidth; i < barWidth; i++ {
		bar += " "
	}

	rpsBits := atomic.LoadUint64(&p.requestsPerSecond)
	rps := math.Float64frombits(rpsBits)
	eta := p.calculateETA()
	
	progressText := fmt.Sprintf("Progress: [%s] %d/%d (%.2f%%) | %.1f req/s | ETA: %s", bar, p.current, p.total, percent, rps, eta)
	p.controller.Overwritef("%s", progressText)
} 