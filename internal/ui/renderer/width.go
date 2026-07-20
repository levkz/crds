package renderer

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

func VisibleWidth(s string) int {
	return runewidth.StringWidth(StripANSI(s))
}

func LineWidth(s string) int {
	return VisibleWidth(s)
}

func MaxLineWidth(text string) int {
	max := 0
	for _, line := range strings.Split(text, "\n") {
		if w := VisibleWidth(line); w > max {
			max = w
		}
	}
	return max
}

func TextDimensions(text string) (width, height int) {
	if text == "" {
		return 0, 0
	}
	lines := strings.Split(text, "\n")
	height = len(lines)
	for _, line := range lines {
		if w := VisibleWidth(line); w > width {
			width = w
		}
	}
	return
}
