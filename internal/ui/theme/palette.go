package theme

import "github.com/charmbracelet/lipgloss"

type Palette struct {
	Blue       lipgloss.Color
	Green      lipgloss.Color
	Orange     lipgloss.Color
	Red        lipgloss.Color
	Gray       lipgloss.Color
	White      lipgloss.Color
	Background lipgloss.Color
	Selection  lipgloss.Color
	Border     lipgloss.Color
	Link       lipgloss.Color
	Surface    lipgloss.Color
}

var DefaultPalette = Palette{
	Blue:       lipgloss.Color("39"),
	Green:      lipgloss.Color("42"),
	Orange:     lipgloss.Color("214"),
	Red:        lipgloss.Color("196"),
	Gray:       lipgloss.Color("240"),
	White:      lipgloss.Color("255"),
	Background: lipgloss.Color("0"),
	Selection:  lipgloss.Color("27"),
	Border:     lipgloss.Color("59"),
	Link:       lipgloss.Color("33"),
	Surface:    lipgloss.Color("235"),
}
