package output

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/morikuni/aec"
)

// TerminalController manages the terminal output, allowing for dynamic lines.
type TerminalController struct {
	writer io.Writer
	mu     sync.Mutex
}

// NewTerminalController creates a new TerminalController.
func NewTerminalController() *TerminalController {
	return &TerminalController{
		writer: os.Stderr,
	}
}

// Printf prints a formatted string to the managed writer.
func (c *TerminalController) Printf(format string, a ...interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fmt.Fprintf(c.writer, format, a...)
}

// Println prints a line to the managed writer.
func (c *TerminalController) Println(a ...interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fmt.Fprintln(c.writer, a...)
}

// Overwritef clears the current line and prints a new formatted string.
func (c *TerminalController) Overwritef(format string, a ...interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Constrói a sequência de controle para subir uma linha e limpar.
	ansiSequence := aec.EmptyBuilder.Up(1).EraseLine(aec.EraseModes.All).ANSI
	
	// Aplica a sequência de controle à string formatada.
	outputString := ansiSequence.Apply(fmt.Sprintf(format, a...))
	
	fmt.Fprint(c.writer, outputString)
} 