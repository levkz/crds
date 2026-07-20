package layout

import "strings"

func Page(header, body, footer string) string {
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n\n")
	b.WriteString(body)
	b.WriteString("\n\n")
	b.WriteString(footer)
	return b.String()
}
