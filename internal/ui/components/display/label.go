package components

import "crds/internal/ui/styles"

func Label(text string) string {
	return styles.MutedText().Render(text)
}
