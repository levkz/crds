package styles

import (
	"crds/internal/ui"
	"crds/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

func FocusedInput() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(ui.Theme.BorderFor(theme.BorderRoleCard)).
		BorderForeground(ui.Theme.Primary.GetForeground()).
		Padding(0, 1)
}
