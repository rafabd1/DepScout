package output

import (
	"fmt"
	"io"
	"os"

	"github.com/morikuni/aec"
)

// TerminalController manages the terminal output, allowing for dynamic lines.
type TerminalController struct {
	writer io.Writer
}

// NewTerminalController creates a new TerminalController.
func NewTerminalController() *TerminalController {
	return &TerminalController{
		writer: os.Stderr,
	}
}

// Printf prints a formatted string to the managed writer.
func (c *TerminalController) Printf(format string, a ...interface{}) {
	fmt.Fprintf(c.writer, format, a...)
}

// Println prints a line to the managed writer.
func (c *TerminalController) Println(a ...interface{}) {
	fmt.Fprintln(c.writer, a...)
}

// Overwritef clears the current line and prints a new formatted string.
func (c *TerminalController) Overwritef(format string, a ...interface{}) {
	// Moves cursor to beginning of line, clears it, then prints.
	outputString := fmt.Sprintf("\r%s%s", aec.EraseLine(aec.EraseModes.All), fmt.Sprintf(format, a...))
	fmt.Fprint(c.writer, outputString)
} 