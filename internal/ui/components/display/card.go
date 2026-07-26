package components

import (
	"strings"

	"crds/internal/ui/renderer"
	"crds/internal/ui/styles"
)

type Card struct {
	Front    string
	Back     []string
	Variants []string
	Notes    string
}

func RenderCard(c Card, revealed bool, width int) string {
	s := styles.Card(width)
	if !revealed {
		return s.Render(c.Front)
	}

	contentWidth := width - 2

	var backLines []string
	for _, item := range c.Back {
		backLines = append(backLines, renderer.Wrap(item, contentWidth)...)
	}
	backContent := strings.Join(backLines, "\n")

	notesContent := strings.Join(renderer.Wrap(c.Notes, contentWidth), "\n")

	return s.Render(
		c.Front +
			"\n\n" +
			backContent +
			"\n\n" +
			notesContent,
	)
}
