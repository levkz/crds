package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func Grid(items []string, cols int) string {
	if len(items) == 0 || cols < 1 {
		return ""
	}

	var rows []string
	for i := 0; i < len(items); i += cols {
		end := i + cols
		if end > len(items) {
			end = len(items)
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, items[i:end]...))
	}
	return strings.Join(rows, "\n")
}
