package ui

import tea "github.com/charmbracelet/bubbletea"

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

// SaveAnswerMsg is emitted by a Quiz screen when a card is graded.
// For flashcard quiz: only CardID, Grade, DeckID are set.
// For typing quiz: UserInput, CorrectAnswer, Similarity are also set.
type SaveAnswerMsg struct {
	DeckID        string
	CardID        string
	Grade         int
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
	Selected []string
}

// BackHandler is implemented by screens that want to handle Back (Esc)
// before the global handler applies default behavior. Return true if
// the screen consumed the event; false to let the global handler proceed.
type BackHandler interface {
	HandleBack() bool
}
