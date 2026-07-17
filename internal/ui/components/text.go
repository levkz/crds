package components

import "crds/internal/ui/styles"

func Text(content string) string {
	return styles.MutedText().Render(content)
}
