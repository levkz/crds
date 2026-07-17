package styles

import (
	"crds/internal/ui"
	"github.com/charmbracelet/lipgloss"
)

func Card(width int) lipgloss.Style {
	return ui.Theme.Primary.
		Width(width).
		Padding(1)
}
