package layout

import "github.com/charmbracelet/lipgloss"

func Center(text string, width int) string {
	return lipgloss.Place(width, 1, lipgloss.Center, lipgloss.Center, text)
}
