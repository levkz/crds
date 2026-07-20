package components

import "crds/internal/ui/styles"

func Header(title string, width int) string {
	return styles.Header(width).Render(title)
}
