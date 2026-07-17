package components

import "crds/internal/ui/styles"

func Header(title string) string {
	return styles.Header(60).Render(title)
}
