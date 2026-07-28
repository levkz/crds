package theme

import "github.com/charmbracelet/lipgloss"

var DarkPalette = Palette{
	Blue:       lipgloss.Color("75"),
	Green:      lipgloss.Color("84"),
	Orange:     lipgloss.Color("215"),
	Red:        lipgloss.Color("203"),
	Gray:       lipgloss.Color("249"),
	White:      lipgloss.Color("255"),
	Background: lipgloss.Color("233"),
	Selection:  lipgloss.Color("27"),
	Border:     lipgloss.Color("236"),
	Link:       lipgloss.Color("39"),
	Surface:    lipgloss.Color("236"),
	Magenta:    lipgloss.Color("177"),
	Purple:     lipgloss.Color("140"),
	Cyan:       lipgloss.Color("117"),
	Yellow:     lipgloss.Color("220"),
	Primary:    lipgloss.Color("75"),
}

var LightPalette = Palette{
	Blue:       lipgloss.Color("27"),
	Green:      lipgloss.Color("34"),
	Orange:     lipgloss.Color("208"),
	Red:        lipgloss.Color("160"),
	Gray:       lipgloss.Color("238"),
	White:      lipgloss.Color("235"),
	Background: lipgloss.Color("255"),
	Selection:  lipgloss.Color("39"),
	Border:     lipgloss.Color("245"),
	Link:       lipgloss.Color("33"),
	Surface:    lipgloss.Color("250"),
	Magenta:    lipgloss.Color("171"),
	Purple:     lipgloss.Color("98"),
	Cyan:       lipgloss.Color("45"),
	Yellow:     lipgloss.Color("208"),
	Primary:    lipgloss.Color("27"),
}

func DarkTheme() Theme {
	return NewTheme(DarkPalette)
}

func LightTheme() Theme {
	return NewTheme(LightPalette)
}

var TokyonightPalette = Palette{
	Blue:       lipgloss.Color("#7aa2f7"),
	Green:      lipgloss.Color("#9ece6a"),
	Orange:     lipgloss.Color("#ff9e64"),
	Red:        lipgloss.Color("#f7768e"),
	Gray:       lipgloss.Color("#7982a8"),
	White:      lipgloss.Color("#c0caf5"),
	Background: lipgloss.Color("#1a1b26"),
	Selection:  lipgloss.Color("#283457"),
	Border:     lipgloss.Color("#15161e"),
	Link:       lipgloss.Color("#2ac3de"),
	Surface:    lipgloss.Color("#16161e"),
	Magenta:    lipgloss.Color("#bb9af7"),
	Purple:     lipgloss.Color("#9d7cd8"),
	Cyan:       lipgloss.Color("#7dcfff"),
	Yellow:     lipgloss.Color("#e0af68"),
	Primary:    lipgloss.Color("#7aa2f7"),
}

var MochaPalette = Palette{
	Blue:       lipgloss.Color("#89b4fa"),
	Green:      lipgloss.Color("#7fd88f"),
	Orange:     lipgloss.Color("#f5a742"),
	Red:        lipgloss.Color("#e06c75"),
	Gray:       lipgloss.Color("#888888"),
	White:      lipgloss.Color("#eeeeee"),
	Background: lipgloss.Color("#0e0e0e"),
	Selection:  lipgloss.Color("#2a2a2a"),
	Border:     lipgloss.Color("#333333"),
	Link:       lipgloss.Color("#56b6c2"),
	Surface:    lipgloss.Color("#181818"),
	Magenta:    lipgloss.Color("#9d7cd8"),
	Purple:     lipgloss.Color("#b48ead"),
	Cyan:       lipgloss.Color("#56b6c2"),
	Yellow:     lipgloss.Color("#e5c07b"),
	Primary:    lipgloss.Color("#fab283"),
}

func MochaTheme() Theme {
	return NewTheme(MochaPalette)
}

func TokyonightTheme() Theme {
	return NewTheme(TokyonightPalette)
}
