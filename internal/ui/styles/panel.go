package styles

import (
	"crds/internal/ui"
	"crds/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

func Panel(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(ui.Theme.BorderFor(theme.BorderRoleContainer)).
		BorderForeground(ui.Theme.Palette.Border).
		Width(width).
		Padding(1, 2)
}
