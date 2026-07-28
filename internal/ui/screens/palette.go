package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"crds/internal/ui"
	components "crds/internal/ui/components/display"
	"crds/internal/ui/keymap"
	"crds/internal/ui/theme"
)

type PaletteModel struct {
	scrollOffset int
	width        int
	height       int
}

func NewPaletteModel() *PaletteModel {
	return &PaletteModel{}
}

func (m *PaletteModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m PaletteModel) Init() tea.Cmd { return nil }

func (m *PaletteModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case keymap.DefaultList.Up.Match(msg):
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
		case keymap.DefaultList.Down.Match(msg):
			m.scrollOffset++
		}
	}
	return m, nil
}

func (m PaletteModel) View() string {
	body := m.buildBody()
	footer := components.Footer(
		keymap.DefaultList.Footer()+" · "+keymap.DefaultGlobal.Back.Help,
		m.width,
	)

	bodyLines := strings.Split(body, "\n")
	footerLines := strings.Count(footer, "\n") + 1

	availHeight := m.height - footerLines - 2
	if availHeight < 1 {
		availHeight = 1
	}

	maxOffset := len(bodyLines) - availHeight
	if maxOffset < 0 {
		maxOffset = 0
		m.scrollOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}

	end := m.scrollOffset + availHeight
	if end > len(bodyLines) {
		end = len(bodyLines)
	}
	visible := strings.Join(bodyLines[m.scrollOffset:end], "\n")

	return visible + "\n\n" + footer
}

func (m PaletteModel) buildBody() string {
	var sections []string
	sections = append(sections, m.buildThemeInfo())
	sections = append(sections, m.buildPaletteSection())
	sections = append(sections, m.buildStylesSection())
	sections = append(sections, m.buildTypographySection())
	sections = append(sections, m.buildIconsSection())
	sections = append(sections, m.buildBordersSection())
	sections = append(sections, m.buildSpacingSection())
	return strings.Join(sections, "\n\n")
}

func (m PaletteModel) buildThemeInfo() string {
	current := theme.CurrentName()
	info := fmt.Sprintf("Current theme: %s  |  Terminal: %dx%d",
		current, m.width, m.height)
	return components.Section("Theme Info", info, m.width)
}

func (m PaletteModel) buildPaletteSection() string {
	p := ui.Theme.Palette
	colors := []struct {
		name    string
		color   lipgloss.Color
		purpose string
	}{
		{"Blue", p.Blue, "Primary interactive elements"},
		{"Green", p.Green, "Success states"},
		{"Orange", p.Orange, "Warning states"},
		{"Red", p.Red, "Danger / destructive"},
		{"Gray", p.Gray, "Secondary / muted text"},
		{"White", p.White, "Primary foreground"},
		{"Background", p.Background, "Canvas / page background"},
		{"Selection", p.Selection, "Selected / focused items"},
		{"Border", p.Border, "Structural dividers"},
		{"Link", p.Link, "Navigable / tappable elements"},
		{"Surface", p.Surface, "Elevated card / container surfaces"},
		{"Magenta", p.Magenta, "Accent / highlight variation"},
		{"Purple", p.Purple, "Accent / highlight variation"},
		{"Cyan", p.Cyan, "Accent / highlight variation"},
		{"Yellow", p.Yellow, "Accent / highlight variation"},
	}

	var lines []string
	for _, c := range colors {
		swatch := lipgloss.NewStyle().Background(c.color).Render("  ")
		name := ui.Theme.Primary.Render(fmt.Sprintf("%-12s", c.name))
		val := ui.Theme.Secondary.Render(fmt.Sprintf("%-10s", `"`+string(c.color)+`"`))
		purpose := ui.Theme.Muted.Render(c.purpose)
		lines = append(lines, "  "+name+"  "+swatch+"  "+val+"  "+purpose)
	}
	return components.Section("Palette", strings.Join(lines, "\n"), m.width)
}

func (m PaletteModel) buildStylesSection() string {
	th := ui.Theme
	styles := []struct {
		name    string
		style   lipgloss.Style
		purpose string
	}{
		{name: "Primary", style: th.Primary, purpose: "Links, active nav items"},
		{name: "Secondary", style: th.Secondary, purpose: "Metadata, timestamps, help text"},
		{name: "Accent", style: th.Accent, purpose: "Emphasis, highlighted terms"},
		{name: "Success", style: th.Success, purpose: "Correct answers, confirmations"},
		{name: "Warning", style: th.Warning, purpose: "Near-limit states, soft errors"},
		{name: "Danger", style: th.Danger, purpose: "Errors, destructive actions"},
		{name: "Muted", style: th.Muted, purpose: "Disabled items, secondary info"},
		{name: "Header", style: th.Header, purpose: "Section headers (bold)"},
		{name: "Background", style: th.Background, purpose: "Content on background surface"},
		{name: "Surface", style: th.Surface, purpose: "Cards, panels, elevated containers"},
		{name: "PrimaryBg", style: th.PrimaryBg, purpose: "Primary button / highlighted bg"},
		{name: "SuccessBg", style: th.SuccessBg, purpose: "Success state background"},
		{name: "ErrorBg", style: th.ErrorBg, purpose: "Error state background"},
		{name: "WarningBg", style: th.WarningBg, purpose: "Warning state background"},
	}

	sample := "The quick brown fox"
	var lines []string
	for _, s := range styles {
		rendered := s.style.Render(sample)
		name := ui.Theme.Primary.Render(fmt.Sprintf("%-12s", s.name))
		purpose := ui.Theme.Muted.Render(s.purpose)
		lines = append(lines, "  "+name+"  "+rendered+"  "+purpose)
	}
	return components.Section("Semantic Styles", strings.Join(lines, "\n"), m.width)
}

func (m PaletteModel) buildTypographySection() string {
	ty := ui.Theme.Typography
	roles := []struct {
		name    string
		style   lipgloss.Style
		purpose string
	}{
		{name: "Title", style: ty.Title, purpose: "Page titles, screen headers"},
		{name: "Subtitle", style: ty.Subtitle, purpose: "Section subtitles, deck names"},
		{name: "Body", style: ty.Body, purpose: "Main content, card text"},
		{name: "Caption", style: ty.Caption, purpose: "Footnotes, timestamps, labels"},
		{name: "Emphasis", style: ty.Emphasis, purpose: "Highlighted terms, important info"},
		{name: "Key", style: ty.Key, purpose: "Keyboard shortcuts, key labels"},
	}

	sample := "The quick brown fox jumps over the lazy dog"
	var lines []string
	for _, r := range roles {
		rendered := r.style.Render(sample)
		name := ui.Theme.Primary.Render(fmt.Sprintf("%-10s", r.name))
		purpose := ui.Theme.Muted.Render(r.purpose)
		lines = append(lines, "  "+name+"  "+rendered+"  "+purpose)
	}
	return components.Section("Typography", strings.Join(lines, "\n"), m.width)
}

func (m PaletteModel) buildIconsSection() string {
	ic := ui.Theme.Icons
	icons := []struct {
		name    string
		glyph   string
		purpose string
	}{
		{name: "Check", glyph: ic.Check, purpose: "Correct answer, confirmed, active state"},
		{name: "Cross", glyph: ic.Cross, purpose: "Incorrect answer, error, inactive"},
		{name: "ArrowUp", glyph: ic.ArrowUp, purpose: "Navigate up, increase"},
		{name: "ArrowDown", glyph: ic.ArrowDown, purpose: "Navigate down, decrease"},
		{name: "ArrowLeft", glyph: ic.ArrowLeft, purpose: "Go back, previous"},
		{name: "Bullet", glyph: ic.Bullet, purpose: "List item marker, decorative separator"},
		{name: "Selected", glyph: ic.Selected, purpose: "Active menu item, current selection marker"},
		{name: "Navigate", glyph: ic.Navigate, purpose: "Forward navigation indicator"},
		{name: "Highlight", glyph: ic.Highlight, purpose: "Featured item, important marker"},
		{name: "Close", glyph: ic.Close, purpose: "Dismiss modal, remove item, clear"},
	}

	var lines []string
	for _, ic := range icons {
		name := ui.Theme.Primary.Render(fmt.Sprintf("%-12s", ic.name))
		glyph := ui.Theme.Accent.Render(" " + ic.glyph + " ")
		purpose := ui.Theme.Muted.Render(ic.purpose)
		lines = append(lines, "  "+name+"  "+glyph+"  "+purpose)
	}
	return components.Section("Icons", strings.Join(lines, "\n"), m.width)
}

func (m PaletteModel) buildBordersSection() string {
	br := ui.Theme.Borders
	borders := []struct {
		name    string
		border  lipgloss.Border
		purpose string
	}{
		{name: "Normal", border: br.Normal, purpose: "Page-level containers, main panels"},
		{name: "Rounded", border: br.Rounded, purpose: "Elevated cards, list items, dialogs"},
		{name: "Double", border: br.Double, purpose: "Important informational blocks"},
		{name: "Thick", border: br.Thick, purpose: "Section dividers, nested groups"},
		{name: "None", border: br.None, purpose: "Clean edges, embedded content"},
	}

	var lines []string
	for _, b := range borders {
		box := lipgloss.NewStyle().
			Border(b.border).
			BorderForeground(ui.Theme.Palette.Border).
			Width(16).
			Padding(0, 1).
			Render(b.name)
		indented := indentLines(box, "  ")
		purpose := ui.Theme.Muted.Render(b.purpose)
		lines = append(lines, indented+"  "+purpose)
	}
	return components.Section("Borders", strings.Join(lines, "\n"), m.width)
}

func (m PaletteModel) buildSpacingSection() string {
	sp := ui.Theme.Spacing
	spacings := []struct {
		name    string
		value   int
		purpose string
	}{
		{name: "Xxs", value: sp.Xxs, purpose: "Tiny gutter, badge padding"},
		{name: "Xs", value: sp.Xs, purpose: "Tight spacing, inline gaps"},
		{name: "Sm", value: sp.Sm, purpose: "Element padding, small gaps"},
		{name: "Md", value: sp.Md, purpose: "Default padding, card margin"},
		{name: "Lg", value: sp.Lg, purpose: "Section spacing"},
		{name: "Xl", value: sp.Xl, purpose: "Major section breaks"},
		{name: "Xxl", value: sp.Xxl, purpose: "Page margins, empty states"},
	}

	var lines []string
	for _, s := range spacings {
		bar := strings.Repeat("█", s.value)
		name := ui.Theme.Primary.Render(fmt.Sprintf("%-5s", s.name))
		val := ui.Theme.Muted.Render(fmt.Sprintf("%3d", s.value))
		barStyled := ui.Theme.Accent.Render(bar)
		purpose := ui.Theme.Muted.Render(s.purpose)
		lines = append(lines, "  "+name+"  "+val+"  "+barStyled+"  "+purpose)
	}
	return components.Section("Spacing (columns)", strings.Join(lines, "\n"), m.width)
}

func indentLines(s, prefix string) string {
	parts := strings.Split(s, "\n")
	for i := range parts {
		parts[i] = prefix + parts[i]
	}
	return strings.Join(parts, "\n")
}
