package styles

import (
	"crds/internal/ui"
	"github.com/charmbracelet/lipgloss"
)

func Error() lipgloss.Style {
	return ui.Theme.Danger
}
