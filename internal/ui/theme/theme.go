package theme

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Palette Palette

	Primary   lipgloss.Style
	Secondary lipgloss.Style
	Accent    lipgloss.Style
	Success   lipgloss.Style
	Warning   lipgloss.Style
	Danger    lipgloss.Style
	Muted     lipgloss.Style
	Header    lipgloss.Style
	Background lipgloss.Style
	Surface   lipgloss.Style

	Typography Typography
	Borders    Borders
	Icons      Icons
	Spacing    Spacing
}

func NewTheme(p Palette) Theme {
	return Theme{
		Palette: p,
		Primary:   lipgloss.NewStyle().Foreground(p.Blue),
		Secondary: lipgloss.NewStyle().Foreground(p.Gray).Faint(true),
		Accent:    lipgloss.NewStyle().Foreground(p.Link),
		Success:   lipgloss.NewStyle().Foreground(p.Green),
		Warning:   lipgloss.NewStyle().Foreground(p.Orange),
		Danger:    lipgloss.NewStyle().Foreground(p.Red),
		Muted:     lipgloss.NewStyle().Foreground(p.Gray),
		Header:    lipgloss.NewStyle().Bold(true).Background(p.Surface),
		Background: lipgloss.NewStyle().Foreground(p.White),
		Surface:   lipgloss.NewStyle().Background(p.Surface).Foreground(p.Blue),

		Typography: NewTypography(p),
		Borders:    DefaultBorders,
		Icons:      DefaultIcons,
		Spacing:    DefaultSpacing,
	}
}

func NewTerminalTheme() Theme {
	th := NewTheme(DefaultPalette)
	th.Icons = DetectedIcons()
	return th
}

func (t Theme) WithIconSource(s IconSource) Theme {
	t.Icons = IconsFromSource(s)
	return t
}

func (t Theme) WithFallbackIcons() Theme {
	t.Icons = FallbackIcons
	return t
}

func (t Theme) BorderFor(role BorderRole) lipgloss.Border {
	switch role {
	case BorderRoleCard, BorderRoleModal:
		return t.Borders.Rounded
	case BorderRoleEmphasis:
		return t.Borders.Double
	case BorderRoleSection:
		return t.Borders.Thick
	case BorderRoleNone:
		return t.Borders.None
	default:
		return t.Borders.Normal
	}
}

var Default = NewTheme(DefaultPalette)
