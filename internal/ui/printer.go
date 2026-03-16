package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	out io.Writer = os.Stdout
	errOut io.Writer = os.Stderr
)

var (
	successColor = color.New(color.FgGreen)
	errorColor   = color.New(color.FgRed)
	warnColor    = color.New(color.FgYellow)
	infoColor    = color.New(color.FgCyan)
	boldColor    = color.New(color.Bold)
	dimColor     = color.New(color.Faint)
)

// Success prints a success message with a green checkmark.
func Success(format string, a ...any) {
	successColor.Fprintf(out, "✓ ")
	fmt.Fprintf(out, format+"\n", a...)
}

// Error prints an error message with a red cross.
func Error(format string, a ...any) {
	errorColor.Fprintf(errOut, "✗ ")
	fmt.Fprintf(errOut, format+"\n", a...)
}

// Warn prints a warning message with a yellow indicator.
func Warn(format string, a ...any) {
	warnColor.Fprintf(errOut, "! ")
	fmt.Fprintf(errOut, format+"\n", a...)
}

// Info prints an informational message with a cyan indicator.
func Info(format string, a ...any) {
	infoColor.Fprintf(out, "→ ")
	fmt.Fprintf(out, format+"\n", a...)
}

// Bold prints bold text.
func Bold(format string, a ...any) {
	boldColor.Fprintf(out, format+"\n", a...)
}

// Dim prints dimmed text.
func Dim(format string, a ...any) {
	dimColor.Fprintf(out, format+"\n", a...)
}

// Table prints a simple aligned table.
func Table(headers []string, rows [][]string) {
	if len(rows) == 0 {
		Dim("No results.")
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	headerLine := make([]string, len(headers))
	separatorLine := make([]string, len(headers))
	for i, h := range headers {
		headerLine[i] = fmt.Sprintf("%-*s", widths[i], h)
		separatorLine[i] = strings.Repeat("─", widths[i])
	}
	boldColor.Fprintf(out, "%s\n", strings.Join(headerLine, "  "))
	dimColor.Fprintf(out, "%s\n", strings.Join(separatorLine, "  "))

	// Print rows
	for _, row := range rows {
		cells := make([]string, len(headers))
		for i := range headers {
			val := ""
			if i < len(row) {
				val = row[i]
			}
			cells[i] = fmt.Sprintf("%-*s", widths[i], val)
		}
		fmt.Fprintf(out, "%s\n", strings.Join(cells, "  "))
	}
}

// KeyValue prints a key-value pair with aligned formatting.
func KeyValue(key, value string) {
	dimColor.Fprintf(out, "%-20s", key)
	fmt.Fprintf(out, "%s\n", value)
}
