package layout

import "github.com/charmbracelet/lipgloss"

func AlignLeft(text string, width int) string {
	return lipgloss.Place(width, 1, lipgloss.Left, lipgloss.Center, text)
}

func AlignRight(text string, width int) string {
	return lipgloss.Place(width, 1, lipgloss.Right, lipgloss.Center, text)
}
