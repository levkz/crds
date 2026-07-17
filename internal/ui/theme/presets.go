package theme

import "github.com/charmbracelet/lipgloss"

var DarkPalette = Palette{
	Blue:       lipgloss.Color("75"),
	Green:      lipgloss.Color("84"),
	Orange:     lipgloss.Color("215"),
	Red:        lipgloss.Color("203"),
	Gray:       lipgloss.Color("242"),
	White:      lipgloss.Color("255"),
	Background: lipgloss.Color("233"),
	Selection:  lipgloss.Color("27"),
	Border:     lipgloss.Color("236"),
	Link:       lipgloss.Color("39"),
	Surface:    lipgloss.Color("236"),
}

var LightPalette = Palette{
	Blue:       lipgloss.Color("27"),
	Green:      lipgloss.Color("34"),
	Orange:     lipgloss.Color("208"),
	Red:        lipgloss.Color("160"),
	Gray:       lipgloss.Color("238"),
	White:      lipgloss.Color("231"),
	Background: lipgloss.Color("255"),
	Selection:  lipgloss.Color("39"),
	Border:     lipgloss.Color("245"),
	Link:       lipgloss.Color("33"),
	Surface:    lipgloss.Color("250"),
}

func DarkTheme() Theme {
	return NewTheme(DarkPalette)
}

func LightTheme() Theme {
	return NewTheme(LightPalette)
}
