package ui

import (
	"os"

	"golang.org/x/term"
)

// getTerminalWidth returns the width of the terminal in columns.
// Returns 0 if the width cannot be determined.
func getTerminalWidth() int {
	fd := int(os.Stderr.Fd())
	w, _, err := term.GetSize(fd)
	if err != nil {
		return 0
	}
	return w
}
