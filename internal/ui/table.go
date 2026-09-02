package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Truncate shortens s to at most max runes, appending "…" when truncated.
func Truncate(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	if max == 1 {
		return "…"
	}
	return string(runes[:max-1]) + "…"
}

// PadRight pads or truncates value to width runes.
func PadRight(value string, width int) string {
	count := utf8.RuneCountInString(value)
	if count > width {
		return Truncate(value, width)
	}
	return value + strings.Repeat(" ", width-count)
}

// ColSpec describes a table column.
type ColSpec struct {
	Value string
	Width int
}

// PrintHeader renders a header row.
func PrintHeader(cols []ColSpec, sep string) string {
	parts := make([]string, len(cols))
	for i, col := range cols {
		parts[i] = PadRight(col.Value, col.Width)
	}
	return strings.Join(parts, sep) + "\n"
}

// DividerLine renders a separator under the header.
func DividerLine(cols []ColSpec, sep string) string {
	parts := make([]string, len(cols))
	for i, col := range cols {
		parts[i] = strings.Repeat("-", col.Width)
	}
	return strings.Join(parts, sep) + "\n"
}

// PrintRow renders a data row.
func PrintRow(cols []ColSpec, sep string) string {
	parts := make([]string, len(cols))
	for i, col := range cols {
		parts[i] = PadRight(col.Value, col.Width)
	}
	return strings.Join(parts, sep) + "\n"
}

// FormatFieldPair prints a simple "Field: Value" line.
func FormatFieldPair(field, value string) string {
	return fmt.Sprintf("%-12s %s\n", field+":", value)
}
