package components

import (
	"strings"

	"crds/internal/ui/styles"
)

func Group(title, content string, width int) string {
	if width < 1 {
		return ""
	}
	var b strings.Builder
	b.WriteString(styles.Hint().Render(title))
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(content)
	return b.String()
}
