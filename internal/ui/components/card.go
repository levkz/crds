package components

import "strings"

type Card struct {
	Front string
	Back  []string
	Notes string
}

func RenderCard(c Card, revealed bool) string {

	if !revealed {
		return cardStyle.Render(c.Front)
	}

	return cardStyle.Render(
		c.Front +
			"\n\n" +
			strings.Join(c.Back, "\n") +
			"\n\n" +
			c.Notes,
	)
}
