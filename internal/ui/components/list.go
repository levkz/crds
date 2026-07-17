package components

import (
	"strings"

	"crds/internal/ui"
	"crds/internal/ui/styles"
)

func RenderList(items []string, selected int) string {
	var b strings.Builder
	for i, item := range items {
		if i > 0 {
			b.WriteString("\n")
		}
		if i == selected {
			b.WriteString(styles.SelectedItem().Render(ui.Theme.Icons.Navigate + " " + item))
		} else {
			b.WriteString("  " + item)
		}
	}
	return b.String()
}
