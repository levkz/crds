package components

import "crds/internal/ui/styles"

func Footer(keys string) string {
	return styles.Footer(60).Render(keys)
}
