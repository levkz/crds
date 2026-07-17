package components

import (
	"strings"

	"crds/internal/ui/styles"
)

type Card struct {
	Front string
	Back  []string
	Notes string
}

func RenderCard(c Card, revealed bool) string {
	s := styles.Card(60)
	if !revealed {
		return s.Render(c.Front)
	}

	return s.Render(
		c.Front +
			"\n\n" +
			strings.Join(c.Back, "\n") +
			"\n\n" +
			c.Notes,
	)
}
