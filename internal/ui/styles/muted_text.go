package styles

import (
	"crds/internal/ui"
	"github.com/charmbracelet/lipgloss"
)

func MutedText() lipgloss.Style {
	return ui.Theme.Muted
}
