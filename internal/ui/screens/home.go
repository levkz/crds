package screens

import (
	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	components "crds/internal/ui/components/display"
	"crds/internal/ui/keymap"
	"crds/internal/ui/layout"
)

type HomeModel struct {
	cursor     int
	activities []string
	width      int
	height     int
}

func NewHome() *HomeModel {
	return &HomeModel{
		activities: []string{
			"Study",
			"Search",
			"Statistics",
			"Settings",
		},
		width:  60,
		height: 24,
	}
}

func (m *HomeModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m HomeModel) Init() tea.Cmd { return nil }

func (m *HomeModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case keymap.DefaultList.Up.Match(msg):
			if m.cursor > 0 {
				m.cursor--
			}
		case keymap.DefaultList.Down.Match(msg):
			if m.cursor < len(m.activities)-1 {
				m.cursor++
			}
		case keymap.DefaultList.Select.Match(msg):
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
	return layout.Page(
		components.Header("Home", m.width),
		components.RenderList(m.activities, m.cursor, m.width),
		components.Footer(keymap.DefaultList.Footer()+" · "+keymap.DefaultGlobal.Help.Help, m.width),
	)
}
