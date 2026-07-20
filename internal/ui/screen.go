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
	SetSize(w, h int)
}

type NavigateToMsg struct {
	Screen ScreenIndex
}

type CardData struct {
	ID    string
	Front string
	Back  []string
	Notes string
}

type DeckData struct {
	Name  string
	Cards []CardData
}

type Stats struct {
	ReviewedToday int
	Accuracy      float64
	TotalCards    int
}

type NavigateToDetailMsg struct {
	Screen ScreenIndex
	Entry  CardData
}

// SaveAnswerMsg is emitted by the Quiz screen when a card is graded.
type SaveAnswerMsg struct {
	CardID string
	Grade  int
}
