package components

import "strings"

func Window(title, content, footer string, width int) string {
	if width < 1 {
		return ""
	}
	var b strings.Builder
	b.WriteString(Header(title, width))
	b.WriteString("\n\n")
	b.WriteString(content)
	b.WriteString("\n\n")
	b.WriteString(Footer(footer, width))
	return b.String()
}
