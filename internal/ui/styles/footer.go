package styles

import (
	"crds/internal/ui"
	"github.com/charmbracelet/lipgloss"
)

func Footer(width int) lipgloss.Style {
	return ui.Theme.Muted.
		Background(ui.Theme.Palette.Surface).
		Width(width).
		Padding(0, 1)
}
