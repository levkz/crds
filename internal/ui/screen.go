package ui

import tea "github.com/charmbracelet/bubbletea"

type ScreenIndex int

const (
	HomeScreen     ScreenIndex = iota
	QuizScreen
	SearchScreen
	StatisticsScreen
	SettingsScreen
	DetailScreen
)

type Screen interface {
	Init() tea.Cmd
	Update(tea.Msg) (Screen, tea.Cmd)
	View() string
}

type NavigateToMsg struct {
	Screen ScreenIndex
}
