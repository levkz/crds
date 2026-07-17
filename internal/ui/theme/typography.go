package theme

import "github.com/charmbracelet/lipgloss"

type Typography struct {
	Title     lipgloss.Style
	Subtitle  lipgloss.Style
	Body      lipgloss.Style
	Caption   lipgloss.Style
	Emphasis  lipgloss.Style
	Key       lipgloss.Style
}

func NewTypography(p Palette) Typography {
	return Typography{
		Title:    lipgloss.NewStyle().Bold(true).Foreground(p.Blue),
		Subtitle: lipgloss.NewStyle().Foreground(p.Gray),
		Body:     lipgloss.NewStyle(),
		Caption:  lipgloss.NewStyle().Italic(true).Foreground(p.Gray),
		Emphasis: lipgloss.NewStyle().Bold(true).Italic(true),
		Key:      lipgloss.NewStyle().Foreground(p.Gray),
	}
}
