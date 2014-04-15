// 2013 Iain Shigeoka - BSD license (see LICENSE)
package cli

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

func NewTerminal(program *Program) *Terminal {
	return &Terminal{Program: program, IndentSize: 2}
}

type Terminal struct {
	Program    *Program // The program this terminal belongs to
	Indent     uint     // Current ident level for stdout statements
	IndentSize uint     // Number of spaces to indent stdout statements
}

// -------------------------------------------
// Identing support for log-style output
// -------------------------------------------

// Increase the indent level
func (t *Terminal) PushIndent() *Terminal {
	t.Indent++
	return t
}

// Decrease the indent level
func (t *Terminal) PopIndent() *Terminal {
	if t.Indent > 0 {
		t.Indent--
	}
	return t
}

// Reduce indent level to zero
func (t *Terminal) PopIndents() *Terminal {
	t.Indent = 0
	return t
}

// -------------------------------------------
// Simple, log-style output
// -------------------------------------------

// Outputs the provided message only if the program is in verbose mode
func (t *Terminal) Verbose(msg string) {
	t.printIndent()
	fmt.Println(msg)
}

// Outputs the provided message only if the program is in verbose mode
func (t *Terminal) Verbosef(format string, data ...interface{}) {
	t.printIndent()
	fmt.Println(fmt.Sprintf(format, data...))
}

// Outputs the provided message
func (t *Terminal) Info(msg string) {
	t.printIndent()
	fmt.Println(msg)
}

// Outputs the provided message
func (t *Terminal) Infof(format string, data ...interface{}) {
	t.printIndent()
	fmt.Printf(format, data...)
	fmt.Println()
}

// Outputs the provided error message and exits the program with an error code
func (t *Terminal) Fatal(msg string) {
	// TODO pretty print the error(s) if exists
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

// Outputs the provided message
func (t *Terminal) Fatalf(format string, data ...interface{}) {
	fmt.Fprintf(os.Stderr, format, data...)
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}

// Outputs the provided error message and exits the program with an error code
// only if the provided error is non-nil. If the program is in verbose mode,
// the error itself is dumped.
func (t *Terminal) Error(err error, msg string) {
	if err != nil {
		if !reflect.ValueOf(err).IsNil() {
			if msg != "" {
				fmt.Fprintln(os.Stderr, msg)
			}
			t.printError(err)
		}
	}
}

// Outputs the provided message
func (t *Terminal) Errorf(err error, format string, data ...interface{}) {
	if err != nil {
		if !reflect.ValueOf(err).IsNil() {
			if format != "" {
				fmt.Fprintf(os.Stderr, format, data...)
				fmt.Fprintln(os.Stderr)
			}
			t.printError(err)
		}
	}
}

// -------------------------------------------
// Chainable i/o
// -------------------------------------------

// Prints characters to the screen (ignores indent and does not append nl)
func (t *Terminal) Print(format string, data ...interface{}) *Terminal {
	if len(data) > 0 {
		fmt.Printf(format, data...)
	} else {
		fmt.Print(format)
	}
	return t
}

func (t *Terminal) Nl(a ...int) *Terminal {
	length := 1
	if len(a) > 0 {
		length = a[0]
	}
	for i := 0; i < length; i++ {
		t.Print("\n")
	}
	return t
}

// -------------------------------------------
// Cursor instructions
// -------------------------------------------

// Clears the entire screen of text and sets the cursor at the top left of the screen.
func (t *Terminal) Clear() *Terminal {
	return t.Print("\033[2J")
}

// Clears the current line of text.
func (t *Terminal) ClearLine() *Terminal {
	return t.Print("\033[2K")
}

// Moves cursor to the absolute coordinates x, y. Values are 1-based and default to top left corner of the screen.
func (t *Terminal) Move(x, y int) *Terminal {
	return t.Print("\033[%d;%dH", x, y)
}

// Moves cursor 'x' cells up. If the edge of the screen is reached, does nothing.
func (t *Terminal) Up(x int) *Terminal {
	return t.Print("\033[%dA", x)
}

// Moves cursor 'x' cells dwn. If the edge of the screen is reached, does nothing.
func (t *Terminal) Down(x int) *Terminal {
	return t.Print("\033[%dB", x)
}

// Moves cursor 'x' cells to the left. If the edge of the screen is reached, does nothing.
func (t *Terminal) Left(x int) *Terminal {
	return t.Print("\033[%dD", x)
}

// Moves cursor 'x' cells to the right. If the edge of the screen is reached, does nothing.
func (t *Terminal) Right(x int) *Terminal {
	return t.Print("\033[%dC", x)
}

// Move the cursor to the beginning of the line "x" lines down.
func (t *Terminal) NextLine(x int) *Terminal {
	return t.Print("\033[%dE", x)
}

// Move the cursor to the beginning of the line "x" lines up.
func (t *Terminal) PreviousLine(x int) *Terminal {
	return t.Print("\033[%dF", x)
}

// Hide the cursor
func (t *Terminal) Hide() *Terminal {
	return t.Print("\033[?25h")
}

// Show the cursor
func (t *Terminal) Show() *Terminal {
	return t.Print("\033[?25l")
}

// -------------------------------------------
// Color
// -------------------------------------------

const (
	Black = iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

func (t *Terminal) Color(foreground, background int) *Terminal {
	return t.Print("\033[3%dm;4%dm;", foreground, background)
}

// Reset terminal attributes (including colors) to default values.
func (t *Terminal) Reset() *Terminal {
	return t.Print("\033[0m")
}

// -------------------------------------------
// Helpers
// -------------------------------------------

// Prints stdout indents if necessary
func (t *Terminal) printIndent() {
	if t.Indent > 0 {
		fmt.Print(strings.Repeat(" ", (int)(t.Indent*t.IndentSize)))
	}
}

// Pretty prints the error in error messages
func (t *Terminal) printError(err error) {
	// if p.verbose {
	fmt.Fprintf(os.Stderr, "\nError: %#v\n", err)
	// }
	// TODO if err has an error code, use that for the exit code
	os.Exit(1)
}
