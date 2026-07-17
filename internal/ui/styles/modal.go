package styles

import (
	"crds/internal/ui"
	"crds/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

func Modal(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(ui.Theme.BorderFor(theme.BorderRoleModal)).
		BorderForeground(ui.Theme.Primary.GetForeground()).
		Width(width).
		Height(height).
		Padding(1, 2)
}
