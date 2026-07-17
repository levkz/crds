package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	"crds/internal/ui/components"
	"crds/internal/ui/theme"
)

type SettingsModel struct {
	cursor int
	themes []string
}

func NewSettings() SettingsModel {
	return SettingsModel{
		themes: theme.Names(),
	}
}

func (m SettingsModel) Init() tea.Cmd { return nil }

func (m SettingsModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.themes)-1 {
				m.cursor++
			}
		case "enter":
			return m, func() tea.Msg {
				return ui.ThemeSwitchMsg{Name: m.themes[m.cursor]}
			}
		}
	}
	return m, nil
}

func (m SettingsModel) View() string {
	current := theme.CurrentName()
	var b strings.Builder
	b.WriteString(components.Header("Settings"))
	b.WriteString("\n\n")
	b.WriteString(components.Text("Select a theme"))
	b.WriteString("\n")
	for i, name := range m.themes {
		marker := "  "
		if i == m.cursor {
			marker = ui.Theme.Primary.Render(ui.Theme.Icons.Navigate)
		}
		line := "  " + marker + " " + name
		if name == current {
			line += " " + ui.Theme.Success.Render(ui.Theme.Icons.Check)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(components.Footer("↑/↓ navigate · enter select · esc back"))
	return b.String()
}
