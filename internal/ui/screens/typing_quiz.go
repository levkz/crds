package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"crds/internal/fuzzy"
	"crds/internal/model"
	"crds/internal/ui"
	components "crds/internal/ui/components/display"
	icomp "crds/internal/ui/components/interactive"
	"crds/internal/ui/keymap"
	"crds/internal/ui/layout"
	"crds/internal/ui/styles"
)

type TypingQuizModel struct {
	cards     []components.Card
	cardIndex int
	progress  int
	revealed  bool
	input     icomp.TextInputModel
	grade     int
	score     float64
	deckName  string
	inverse   bool
	width     int
	height    int
	matcher   *fuzzy.FuzzyMatcher
}

func NewTypingQuiz() *TypingQuizModel {
	input := icomp.NewTextInput()
	input.Focus()
	return &TypingQuizModel{
		input:   input,
		matcher: fuzzy.NewFuzzyMatcher(0),
	}
}

func (m *TypingQuizModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *TypingQuizModel) SetDeck(deck ui.DeckData) {
	cards := make([]components.Card, len(deck.Cards))
	for i, c := range deck.Cards {
		cards[i] = components.Card{
			Front: c.Front,
			Back:  c.Back,
			Variants: func() []string {
				if len(c.Variants) > 0 {
					return c.Variants
				}
				return c.Back
			}(),
			Notes: c.Notes,
		}
	}
	m.cards = cards
	m.deckName = deck.Name
	m.cardIndex = 0
	m.progress = 0
	m.revealed = false
	m.inverse = false
	m.grade = 0
	m.score = 0
	m.input.SetValue("")
}

func (m *TypingQuizModel) Init() tea.Cmd { return nil }

func (m *TypingQuizModel) currentVariants() []string {
	if m.inverse {
		return model.ExpandText(m.cards[m.cardIndex].Front)
	}
	return m.cards[m.cardIndex].Variants
}

func (m *TypingQuizModel) currentCorrectAnswer() string {
	if m.inverse {
		return m.cards[m.cardIndex].Front
	}
	return strings.Join(m.cards[m.cardIndex].Back, ", ")
}

func (m *TypingQuizModel) gradeInput() bool {
	answer := m.input.Value()
	if answer == "" {
		return false
	}
	variants := m.currentVariants()
	m.grade = m.matcher.Grade(answer, variants)
	score, _ := m.matcher.Check(answer, variants)
	m.score = score
	return true
}

func (m *TypingQuizModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case keymap.DefaultGlobal.Back.Match(msg):
			return m, func() tea.Msg {
				return ui.NavigateToMsg{Screen: ui.HomeScreen}
			}

		case !m.revealed:
			switch {
			case keymap.DefaultTypingQuiz.Submit.Match(msg):
				if m.gradeInput() {
					m.revealed = true
					card := m.cards[m.cardIndex]
				return m, func() tea.Msg {
					return ui.SaveAnswerMsg{
						DeckID:        m.deckName,
						CardID:        card.Front,
						Grade:         m.grade,
						Reverse:       m.inverse,
						UserInput:     m.input.Value(),
						CorrectAnswer: m.currentCorrectAnswer(),
						Similarity:    m.score,
					}
				}
				}
				return m, nil

			case keymap.DefaultTypingQuiz.Reveal.Match(msg):
				m.revealed = true
				m.grade = fuzzy.Again
				m.score = 0
				card := m.cards[m.cardIndex]
				return m, func() tea.Msg {
					return ui.SaveAnswerMsg{
						DeckID:  m.deckName,
						CardID:  card.Front,
						Grade:   fuzzy.Again,
						Reverse: m.inverse,
					}
				}

			case keymap.DefaultTypingQuiz.Inverse.Match(msg):
				m.inverse = !m.inverse
				m.revealed = false
				m.grade = 0
				m.score = 0
				m.input.SetValue("")

			default:
				updated, cmd := m.input.Update(msg)
				m.input = updated
				return m, cmd
			}

		default:
			switch {
			case keymap.DefaultTypingQuiz.Submit.Match(msg),
				keymap.DefaultList.Down.Match(msg):
				m.progress = (m.cardIndex + 1) * 100 / len(m.cards)
				m.cardIndex++
				if m.cardIndex >= len(m.cards) {
					return m, func() tea.Msg {
						return ui.NavigateToMsg{Screen: ui.StatisticsScreen}
					}
				}
				m.revealed = false
				m.grade = 0
				m.score = 0
				m.input.SetValue("")
				return m, nil
			}
		}
	}

	updated, cmd := m.input.Update(msg)
	m.input = updated
	return m, cmd
}

func (m TypingQuizModel) View() string {
	if len(m.cards) == 0 {
		return layout.Page(
			components.Header("Typing Quiz", m.width),
			styles.MutedText().Render("No cards loaded"),
			components.Footer(keymap.DefaultGlobal.Back.Help, m.width),
			m.height,
		)
	}

	if m.cardIndex >= len(m.cards) {
		return layout.Page(
			components.Header(m.deckName, m.width),
			styles.MutedText().Render("Quiz complete!"),
			components.Footer(keymap.DefaultGlobal.Back.Help, m.width),
			m.height,
		)
	}

	headerTitle := m.deckName
	if headerTitle == "" {
		headerTitle = "Typing Quiz"
	}
	if m.inverse {
		headerTitle += " (inverse)"
	}

	card := m.cards[m.cardIndex]
	displayCard := card
	if m.inverse {
		displayCard = components.Card{
			Front: strings.Join(card.Back, ", "),
			Back:  []string{card.Front},
			Notes: card.Notes,
		}
	}

	var items []string
	items = append(items, components.RenderCard(displayCard, m.revealed, m.width))
	items = append(items, components.ProgressBar(m.progress))
	items = append(items, m.input.View(m.width))

	if m.revealed {
		correctAnswer := m.currentCorrectAnswer()

		var gradeIndicator string
		switch m.grade {
		case fuzzy.Good:
			gradeIndicator = styles.Success().Render("✓ Correct!")
		case fuzzy.Hard:
			gradeIndicator = styles.Warning().Render("~ Close!")
		case fuzzy.Again:
			gradeIndicator = styles.Error().Render("✗ Not quite.")
		}

		var answerLine string
		switch m.grade {
		case fuzzy.Good:
			answerLine = styles.Success().Render("✓ " + m.input.Value())
		case fuzzy.Hard:
			pct := int(m.score * 100)
			answerLine = styles.Warning().Render(fmt.Sprintf("~ %s (%d%%)", m.input.Value(), pct))
		case fuzzy.Again:
			answerLine = styles.Error().Render("✗ " + m.input.Value())
		}

		correctLine := styles.MutedText().Render("Correct: " + correctAnswer)

		items = append(items, layout.Column(gradeIndicator, correctLine, answerLine))
	}

	var footer string
	if m.revealed {
		footer = components.Footer("enter/down next · esc back", m.width)
	} else {
		footer = components.Footer(
			keymap.DefaultTypingQuiz.Footer()+" · "+keymap.DefaultGlobal.Back.Help,
			m.width,
		)
	}

	return layout.Page(
		components.Header(headerTitle, m.width),
		layout.Column(items...),
		footer,
		m.height,
	)
}
