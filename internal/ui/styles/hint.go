package styles

import (
	"crds/internal/ui"
	"github.com/charmbracelet/lipgloss"
)

func Hint() lipgloss.Style {
	return ui.Theme.Muted.Italic(true)
}
