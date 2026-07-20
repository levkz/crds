package screens

import (
	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	components "crds/internal/ui/components/display"
	"crds/internal/ui/events"
	"crds/internal/ui/keymap"
	"crds/internal/ui/layout"
	"crds/internal/ui/theme"
)

type SettingsModel struct {
	cursor int
	themes []string
	width  int
	height int
}

func NewSettings() *SettingsModel {
	return &SettingsModel{
		themes: theme.Names(),
		width:  60,
		height: 24,
	}
}

func (m *SettingsModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m SettingsModel) Init() tea.Cmd { return nil }

func (m *SettingsModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case keymap.DefaultList.Up.Match(msg):
			if m.cursor > 0 {
				m.cursor--
			}
		case keymap.DefaultList.Down.Match(msg):
			if m.cursor < len(m.themes)-1 {
				m.cursor++
			}
			case keymap.DefaultList.Select.Match(msg):
			return m, func() tea.Msg {
				return events.ThemeSwitchMsg{Name: m.themes[m.cursor]}
			}
		}
	}
	return m, nil
}

func (m SettingsModel) View() string {
	current := theme.CurrentName()
	items := make([]string, 0, len(m.themes)+1)
	items = append(items, components.Text("Select a theme"))
	for i, name := range m.themes {
		marker := "  "
		if i == m.cursor {
			marker = ui.Theme.Primary.Render(ui.Theme.Icons.Navigate)
		}
		line := "  " + marker + " " + name
		if name == current {
			line += " " + ui.Theme.Success.Render(ui.Theme.Icons.Check)
		}
		items = append(items, line)
	}
	return layout.Page(
		components.Header("Settings", m.width),
		layout.Column(items...),
		components.Footer(keymap.DefaultList.Footer()+" · "+keymap.DefaultGlobal.Back.Help, m.width),
	)
}
