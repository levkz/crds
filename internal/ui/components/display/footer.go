package components

import "crds/internal/ui/styles"

func Footer(keys string, width int) string {
	return styles.Footer(width).Render(keys)
}
