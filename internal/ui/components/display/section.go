package components

import (
	"strings"

	"crds/internal/ui/renderer"
	"crds/internal/ui/styles"
)

func Section(title, content string, width int) string {
	if width < 1 {
		return ""
	}
	titleLine := "─── " + title + " "
	remaining := width - renderer.VisibleWidth(titleLine)
	if remaining < 1 {
		remaining = 1
	}
	titleLine += strings.Repeat("─", remaining)
	var b strings.Builder
	b.WriteString(styles.MutedText().Render(titleLine))
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(content)
	return b.String()
}
