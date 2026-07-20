package components

import (
	"strings"

	"crds/internal/ui/renderer"
)

func Paragraph(content string, width int) string {
	if content == "" || width < 1 {
		return ""
	}
	return strings.Join(renderer.Wrap(content, width), "\n")
}
