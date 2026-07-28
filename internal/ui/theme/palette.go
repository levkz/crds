package theme

import "github.com/charmbracelet/lipgloss"

type Palette struct {
	Blue, Green, Orange, Red, Gray, White       lipgloss.Color
	Background, Selection, Border, Link, Surface lipgloss.Color
	Magenta, Purple, Cyan, Yellow               lipgloss.Color
	Primary, Secondary, Accent                   lipgloss.Color
}

var DefaultPalette = Palette{
	Blue:       lipgloss.Color("39"),
	Green:      lipgloss.Color("42"),
	Orange:     lipgloss.Color("214"),
	Red:        lipgloss.Color("196"),
	Gray:       lipgloss.Color("248"),
	White:      lipgloss.Color("255"),
	Background: lipgloss.Color("0"),
	Selection:  lipgloss.Color("27"),
	Border:     lipgloss.Color("59"),
	Link:       lipgloss.Color("33"),
	Surface:    lipgloss.Color("235"),
	Magenta:    lipgloss.Color("177"),
	Purple:     lipgloss.Color("140"),
	Cyan:       lipgloss.Color("117"),
	Yellow:     lipgloss.Color("220"),
}
