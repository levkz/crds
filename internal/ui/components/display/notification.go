package components

import "crds/internal/ui/styles"

func RenderNotification(text string) string {
	return styles.Hint().Render(text)
}
