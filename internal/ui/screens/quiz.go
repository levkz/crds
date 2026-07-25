package screens

import (
	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	components "crds/internal/ui/components/display"
	"crds/internal/ui/keymap"
	"crds/internal/ui/layout"
	"crds/internal/ui/styles"
)

type QuizModel struct {
	CardIndex int
	Revealed  bool
	Progress  int
	Cards     []components.Card
	deckName  string

	width  int
	height int
}

func NewQuiz() *QuizModel {
	return &QuizModel{}
}

func (m *QuizModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m QuizModel) Init() tea.Cmd { return nil }

type Grade int

const (
	Again Grade = iota
	Hard
	Good
	Easy
)

func (m *QuizModel) SetDeck(deck ui.DeckData) {
	cards := make([]components.Card, len(deck.Cards))
	for i, c := range deck.Cards {
		cards[i] = components.Card{
			Front: c.Front,
			Back:  c.Back,
			Notes: c.Notes,
		}
	}
	m.Cards = cards
	m.deckName = deck.Name
	m.CardIndex = 0
	m.Progress = 0
	m.Revealed = false
}

func (m *QuizModel) grade(g Grade) (*QuizModel, tea.Cmd) {
	if m.CardIndex >= len(m.Cards) {
		return m, nil
	}

	card := m.Cards[m.CardIndex]
	cardID := card.Front

	m.Progress = (m.CardIndex + 1) * 100 / len(m.Cards)
	m.CardIndex++
	m.Revealed = false

	if m.CardIndex >= len(m.Cards) {
		return m, tea.Sequence(
			func() tea.Msg {
				return ui.NavigateToMsg{Screen: ui.StatisticsScreen}
			},
		)
	}

	return m, func() tea.Msg {
		return ui.SaveAnswerMsg{
			DeckID: m.deckName,
			CardID: cardID,
			Grade:  int(g),
		}
	}
}

func (m *QuizModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {
		case keymap.DefaultQuiz.Reveal.Match(msg):
			m.Revealed = true

		case keymap.DefaultQuiz.Again.Match(msg):
			updated, cmd := m.grade(Again)
			return updated, cmd

		case keymap.DefaultQuiz.Hard.Match(msg):
			updated, cmd := m.grade(Hard)
			return updated, cmd

		case keymap.DefaultQuiz.Good.Match(msg):
			updated, cmd := m.grade(Good)
			return updated, cmd

		case keymap.DefaultQuiz.Easy.Match(msg):
			updated, cmd := m.grade(Easy)
			return updated, cmd
		}
	}

	return m, nil
}

func (m QuizModel) View() string {
	if len(m.Cards) == 0 {
		return layout.Page(
			components.Header("Quiz", m.width),
			styles.MutedText().Render("No cards loaded"),
			components.Footer(keymap.DefaultGlobal.Back.Help, m.width),
			m.height,
		)
	}

	if m.CardIndex >= len(m.Cards) {
		return layout.Page(
			components.Header(m.deckName, m.width),
			styles.MutedText().Render("Quiz complete!"),
			components.Footer(keymap.DefaultGlobal.Back.Help, m.width),
			m.height,
		)
	}

	footer := components.Footer(keymap.DefaultQuiz.Unrevealed(), m.width)
	if m.Revealed {
		footer = components.Footer(keymap.DefaultQuiz.Revealed(), m.width)
	}

	headerTitle := m.deckName
	if headerTitle == "" {
		headerTitle = "Quiz"
	}
	return layout.Page(
		components.Header(headerTitle, m.width),
		layout.Column(
			components.RenderCard(m.Cards[m.CardIndex], m.Revealed, m.width),
			components.ProgressBar(m.Progress),
		),
		footer,
		m.height,
	)
}


