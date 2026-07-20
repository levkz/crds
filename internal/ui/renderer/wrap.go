package renderer

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

func Wrap(text string, maxWidth int) []string {
	if text == "" {
		return nil
	}
	if maxWidth < 1 {
		maxWidth = 1
	}

	var lines []string
	for _, line := range strings.Split(text, "\n") {
		if VisibleWidth(line) <= maxWidth {
			lines = append(lines, line)
			continue
		}

		words := strings.Fields(line)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		var cur strings.Builder
		for _, word := range words {
			ww := VisibleWidth(word)
			if ww > maxWidth {
				if cur.Len() > 0 {
					lines = append(lines, cur.String())
					cur.Reset()
				}
				parts := hardWrap(word, maxWidth)
				lines = append(lines, parts[:len(parts)-1]...)
				cur.WriteString(parts[len(parts)-1])
				continue
			}

			cw := VisibleWidth(cur.String())
			if cur.Len() > 0 && cw+1+ww > maxWidth {
				lines = append(lines, cur.String())
				cur.Reset()
			}

			if cur.Len() > 0 {
				cur.WriteString(" ")
			}
			cur.WriteString(word)
		}

		if cur.Len() > 0 {
			lines = append(lines, cur.String())
		}
	}
	return lines
}

func hardWrap(word string, maxWidth int) []string {
	var parts []string
	var cur strings.Builder
	cw := 0
	for _, r := range word {
		rw := runewidth.RuneWidth(r)
		if cw+rw > maxWidth && cur.Len() > 0 {
			parts = append(parts, cur.String())
			cur.Reset()
			cw = 0
		}
		cur.WriteRune(r)
		cw += rw
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	return parts
}

func Truncate(text string, maxWidth int) string {
	if maxWidth < 1 {
		return ""
	}
	if VisibleWidth(text) <= maxWidth {
		return text
	}
	return runewidth.Truncate(StripANSI(text), maxWidth, "…")
}

func Fit(text string, maxWidth int) string {
	if maxWidth < 1 {
		return ""
	}
	vw := VisibleWidth(text)
	if vw <= maxWidth {
		return text
	}
	return runewidth.Truncate(StripANSI(text), maxWidth, "")
}
