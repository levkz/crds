package components

import (
	"strings"

	"crds/internal/ui"
	"crds/internal/ui/renderer"
	"crds/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

func StatusBar(left, right string, width int, bg lipgloss.Color) string {
	if width < 1 {
		return ""
	}
	leftWidth := renderer.VisibleWidth(left)
	rightWidth := renderer.VisibleWidth(right)
	gap := width - leftWidth - rightWidth
	if gap < 1 {
		gap = 1
	}

	padding := strings.Repeat(" ", gap)

	if bg == "" {
		return styles.MutedText().Width(width).Render(left + padding + right)
	}

	leftRendered := styles.MutedText().Background(bg).Render(left)
	gapRendered := lipgloss.NewStyle().Background(bg).Render(padding)
	rightRendered := ui.Theme.Primary.Background(bg).Render(right)
	return leftRendered + gapRendered + rightRendered
}
