package styles

import (
	"crds/internal/ui"
	"github.com/charmbracelet/lipgloss"
)

func Header(width int) lipgloss.Style {
	return ui.Theme.Header.
		Width(width).
		Padding(0, 1)
}
