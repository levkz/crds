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
	PrimaryBg lipgloss.Style
	SuccessBg lipgloss.Style
	ErrorBg   lipgloss.Style
	WarningBg lipgloss.Style

	Typography Typography
	Borders    Borders
	Icons      Icons
	Spacing    Spacing
}

func primaryOf(p Palette) lipgloss.Color {
	if p.Primary != "" {
		return p.Primary
	}
	return p.Blue
}

func secondaryOf(p Palette) lipgloss.Color {
	if p.Secondary != "" {
		return p.Secondary
	}
	return p.Cyan
}

func accentOf(p Palette) lipgloss.Color {
	if p.Accent != "" {
		return p.Accent
	}
	return p.Orange
}

func (p Palette) bgStyle(fg, bg lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Background(bg).Foreground(fg)
}

func NewTheme(p Palette) Theme {
	primary := primaryOf(p)
	secondary := secondaryOf(p)
	accent := accentOf(p)

	return Theme{
		Palette:   p,
		Primary:   lipgloss.NewStyle().Foreground(primary),
		Secondary: lipgloss.NewStyle().Foreground(secondary),
		Accent:    lipgloss.NewStyle().Foreground(accent),
		Success:   lipgloss.NewStyle().Foreground(p.Green),
		Warning:   lipgloss.NewStyle().Foreground(p.Orange),
		Danger:    lipgloss.NewStyle().Foreground(p.Red),
		Muted:     lipgloss.NewStyle().Foreground(p.Gray),
		Header:    lipgloss.NewStyle().Bold(true).Background(p.Surface).Foreground(p.White),
		Background: lipgloss.NewStyle().Background(p.Background).Foreground(p.White),
		Surface:   lipgloss.NewStyle().Background(p.Surface).Foreground(primary),
		PrimaryBg: p.bgStyle(p.Background, primary),
		SuccessBg: p.bgStyle(p.Background, p.Green),
		ErrorBg:   p.bgStyle(p.Background, p.Red),
		WarningBg: p.bgStyle(p.Background, p.Orange),

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

var Default = parseBuiltin("default")
