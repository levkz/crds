package layout

import "github.com/charmbracelet/lipgloss"

func Row(items ...string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, items...)
}
