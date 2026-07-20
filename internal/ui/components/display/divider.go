package components

import (
	"strings"

	"crds/internal/ui/styles"
)

func Divider(width int) string {
	if width < 1 {
		return ""
	}
	return styles.MutedText().Render(strings.Repeat("─", width))
}
