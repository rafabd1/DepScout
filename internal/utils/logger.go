package utils

import (
	"fmt"
	"time"

	"DepScout/internal/output"

	"github.com/morikuni/aec"
)

// Logger provides structured logging capabilities.
type Logger struct {
	controller *output.TerminalController
	verbose    bool
}

// NewLogger creates a new Logger.
func NewLogger(controller *output.TerminalController, verbose bool) *Logger {
	return &Logger{
		controller: controller,
		verbose:    verbose,
	}
}

func (l *Logger) log(level, color, format string, a ...interface{}) {
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

// Warnf logs a warning message.
func (l *Logger) Warnf(format string, a ...interface{}) {
	l.log("WARN", "yellow", format, a...)
}

// Errorf logs an error message.
func (l *Logger) Errorf(format string, a ...interface{}) {
	l.log("ERROR", "red", format, a...)
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