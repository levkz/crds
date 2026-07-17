package styles

import (
	"crds/internal/ui"
	"github.com/charmbracelet/lipgloss"
)

func SelectedItem() lipgloss.Style {
	return ui.Theme.Primary.
		Background(ui.Theme.Palette.Selection)
}
