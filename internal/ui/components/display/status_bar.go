package components

import (
	"crds/internal/ui/renderer"
	"crds/internal/ui/styles"
)

func StatusBar(left, right string, width int) string {
	if width < 1 {
		return ""
	}
	leftWidth := renderer.VisibleWidth(left)
	rightWidth := renderer.VisibleWidth(right)
	gap := width - leftWidth - rightWidth
	if gap < 1 {
		gap = 1
	}
	padding := ""
	for i := 0; i < gap; i++ {
		padding += " "
	}
	return styles.MutedText().Width(width).Render(left + padding + right)
}
