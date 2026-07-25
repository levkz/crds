package layout

import "strings"

func Page(header, body, footer string, height int) string {
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n\n")
	b.WriteString(body)

	if height > 0 {
		// totalLines = currentLines + 1 + footerLines because
		// the first line of the footer sits on the last blank line
		// created by the trailing "\n\n" separator.
		currentLines := lines(b.String())
		totalLines := currentLines + 1 + lines(footer)
		if pad := height - totalLines; pad > 0 {
			b.WriteString(strings.Repeat("\n", pad))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(footer)
	return b.String()
}

func lines(s string) int {
	if s == "" {
		return 0
	}
	return len(strings.Split(s, "\n"))
}
