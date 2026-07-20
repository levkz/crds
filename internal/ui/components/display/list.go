package components

import (
	"strings"

	"crds/internal/ui"
	"crds/internal/ui/renderer"
	"crds/internal/ui/styles"
)

func RenderList(items []string, selected int, width int) string {
	maxItemWidth := width - 3
	if maxItemWidth < 1 {
		maxItemWidth = 1
	}

	var b strings.Builder
	for i, item := range items {
		if i > 0 {
			b.WriteString("\n")
		}
		truncated := renderer.Truncate(item, maxItemWidth)
		if i == selected {
			b.WriteString(styles.SelectedItem().Render(ui.Theme.Icons.Navigate + " " + truncated))
		} else {
			b.WriteString("  ")
			b.WriteString(truncated)
		}
	}
	return b.String()
}
