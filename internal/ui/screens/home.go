package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	"crds/internal/ui/components"
)

type HomeModel struct {
	cursor     int
	activities []string
}

func NewHome() HomeModel {
	return HomeModel{
		activities: []string{
			"Study",
			"Search",
			"Statistics",
			"Settings",
		},
	}
}

func (m HomeModel) Init() tea.Cmd { return nil }

func (m HomeModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.activities)-1 {
				m.cursor++
			}
		case "enter":
			screens := []ui.ScreenIndex{
				ui.QuizScreen,
				ui.SearchScreen,
				ui.StatisticsScreen,
				ui.SettingsScreen,
			}
			return m, func() tea.Msg {
				return ui.NavigateToMsg{Screen: screens[m.cursor]}
			}
		}
	}
	return m, nil
}

func (m HomeModel) View() string {
	var b strings.Builder
	b.WriteString(components.Header("Home"))
	b.WriteString("\n\n")
	b.WriteString(components.RenderList(m.activities, m.cursor))
	b.WriteString("\n\n")
	b.WriteString(components.Footer("↑/↓ navigate · enter select · ? help"))
	return b.String()
}
