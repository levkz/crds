package ui

import (
	"crds/internal/stats"
	tea "github.com/charmbracelet/bubbletea"
)

type ScreenIndex int

const (
	HomeScreen     ScreenIndex = iota
	QuizScreen
	DecksScreen
	TypingQuizScreen
	SearchScreen
	StatisticsScreen
	SettingsScreen
	DetailScreen
	PaletteScreen
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

type ExampleData struct {
	Text        string
	Translation string
}

type CardData struct {
	ID       string
	DeckID   string
	Front    string
	Back     []string
	Variants []string
	Notes    string
	Tags     []string
	Examples []ExampleData
}

type DeckData struct {
	Name  string
	Cards []CardData
}

type Stats = stats.Summary

type NavigateToDetailMsg struct {
	Screen ScreenIndex
	Entry  CardData
}

// SaveAnswerMsg is emitted by a Quiz screen when a card is graded.
// For flashcard quiz: only CardID, Grade, DeckID are set.
// For typing quiz: UserInput, CorrectAnswer, Similarity are also set.
type SaveAnswerMsg struct {
	DeckID        string
	CardID        string
	Grade         Grade
	Reverse       bool
	UserInput     string
	CorrectAnswer string
	Similarity    float64
}

// TypeAnswerMsg is emitted by the TypingQuiz screen when the user submits a typed answer.
type TypeAnswerMsg struct {
	EntryID string
	Answer  string
}

// TypingGradeMsg carries the result of grading a typed answer.
type TypingGradeMsg struct {
	EntryID string
	Grade   int     // 1=Again, 2=Hard, 3=Good
	Score   float64 // similarity score 0.0-1.0
	Correct bool
}

// DeckSelectionChangedMsg is emitted by the Decks screen when the user confirms a new deck selection.
type DeckSelectionChangedMsg struct {
	Selected     []string
	SelectedTags []string
}

// RefreshWordStatsMsg requests per-word statistics for a single entry. Emitted
// by the Statistics screen when a word is selected; the root fetches the data
// and forwards the result back to the active screen.
type RefreshWordStatsMsg struct {
	EntryID string
}

// WordStatsLoadedMsg carries the per-word statistics result. Handled by the
// Statistics screen (screen-local data, not stored in AppState).
type WordStatsLoadedMsg struct {
	EntryID string
	Stats   stats.WordStats
	History []stats.DayPoint
}

// BackHandler is implemented by screens that want to handle Back (Esc)
// before the global handler applies default behavior. Return true if
// the screen consumed the event; false to let the global handler proceed.
type BackHandler interface {
	HandleBack() bool
}

// QuizInProgressChecker is implemented by quiz screens to report whether
// the user is actively answering cards (not yet complete).
type QuizInProgressChecker interface {
	IsInProgress() bool
}
