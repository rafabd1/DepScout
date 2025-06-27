package utils

import (
	"fmt"
	"sync"
	"time"

	"DepScout/internal/output"

	"github.com/morikuni/aec"
)

// Logger provides structured logging capabilities.
type Logger struct {
	controller  *output.TerminalController
	progBar     *output.ProgressBar
	verbose     bool
	isProgBarOn bool
	mu          sync.Mutex
}

// NewLogger creates a new Logger.
func NewLogger(controller *output.TerminalController, verbose bool) *Logger {
	return &Logger{
		controller: controller,
		verbose:    verbose,
	}
}

// SetProgressBar injects the progress bar dependency.
func (l *Logger) SetProgressBar(p *output.ProgressBar) {
	l.progBar = p
	l.progBar.SetMutex(&l.mu)
}

// SetProgBarActive notifies the logger that the progress bar is running.
func (l *Logger) SetProgBarActive(isActive bool) {
	l.isProgBarOn = isActive
}

func (l *Logger) log(level, color, format string, a ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.progBar != nil && l.isProgBarOn {
		l.progBar.Clear()
	}

	prefix := aec.LightBlackF.Apply(fmt.Sprintf("[%s] ", time.Now().Format("15:04:05")))
	levelStr := fmt.Sprintf("[%s] ", level)
	
	var coloredLevel string
	switch color {
	case "green":
		coloredLevel = aec.GreenF.Apply(levelStr)
	case "yellow":
		coloredLevel = aec.YellowF.Apply(levelStr)
	case "red":
		coloredLevel = aec.RedF.Apply(levelStr)
	case "blue":
		coloredLevel = aec.BlueF.Apply(levelStr)
	default:
		coloredLevel = aec.WhiteF.Apply(levelStr)
	}

	message := fmt.Sprintf(format, a...)
	l.controller.Println(prefix + coloredLevel + message)

	if l.progBar != nil && l.isProgBarOn {
		l.progBar.UnsafeRender()
	}
}

// Infof logs an informational message.
func (l *Logger) Infof(format string, a ...interface{}) {
	l.log("INFO", "blue", format, a...)
}

// Debugf logs a debug message only if verbose mode is enabled.
func (l *Logger) Debugf(format string, a ...interface{}) {
	if l.verbose {
		l.log("DEBUG", "white", format, a...)
	}
}

// Warnf logs a warning message only if verbose mode is enabled.
func (l *Logger) Warnf(format string, a ...interface{}) {
	if l.verbose {
		l.log("WARN", "yellow", format, a...)
	}
}

// Errorf logs an error message only if verbose mode is enabled.
func (l *Logger) Errorf(format string, a ...interface{}) {
	if l.verbose {
		l.log("ERROR", "red", format, a...)
	}
}

// Fatalf logs an error message and exits.
func (l *Logger) Fatalf(format string, a ...interface{}) {
	l.log("FATAL", "red", format, a...)
	// In a real scenario, you might want to os.Exit(1) here,
	// but that can complicate testing. Let's rely on the caller to exit.
}

// Successf logs a success message.
func (l *Logger) Successf(format string, a ...interface{}) {
	l.log("SUCCESS", "green", format, a...)
} 