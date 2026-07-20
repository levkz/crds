package components

import "crds/internal/ui/styles"

func Panel(content string, width int) string {
	return styles.Panel(width).Render(content)
}
