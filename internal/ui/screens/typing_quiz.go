package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"crds/internal/fuzzy"
	"crds/internal/model"
	"crds/internal/ui"
	"crds/internal/ui/keymap"
	"crds/internal/ui/layout"
	"crds/internal/ui/styles"
)

type TypingQuizModel struct {
	cards        []ui.CardData
	cardIndex    int
	revealed     bool
	input        string
	cursor       int
	grade        int
	score        float64
	deckName     string
	inverse      bool
	width        int
	height       int
	matcher      *fuzzy.FuzzyMatcher
	examplesPage int
}

func NewTypingQuiz() *TypingQuizModel {
	return &TypingQuizModel{
		matcher: fuzzy.NewFuzzyMatcher(0),
	}
}

func (m *TypingQuizModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *TypingQuizModel) SetDeck(deck ui.DeckData) {
	m.cards = deck.Cards
	m.deckName = deck.Name
	m.cardIndex = 0
	m.revealed = false
	m.inverse = false
	m.grade = 0
	m.score = 0
	m.input = ""
	m.cursor = 0
	m.examplesPage = 0
}

func (m *TypingQuizModel) Init() tea.Cmd { return nil }

func (m *TypingQuizModel) IsInProgress() bool {
	return m.cardIndex > 0 && m.cardIndex < len(m.cards)
}

func (m *TypingQuizModel) currentVariants() []string {
	if m.inverse {
		return model.ExpandText(m.cards[m.cardIndex].Front)
	}
	c := m.cards[m.cardIndex]
	if len(c.Variants) > 0 {
		return c.Variants
	}
	return c.Back
}

func (m *TypingQuizModel) currentCorrectAnswer() string {
	if m.inverse {
		return m.cards[m.cardIndex].Front
	}
	return strings.Join(m.cards[m.cardIndex].Back, ", ")
}

func (m *TypingQuizModel) gradeInput() bool {
	answer := m.input
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

		case len(m.cards) > 0 && m.cardIndex >= len(m.cards):
			if keymap.DefaultTypingQuiz.Submit.Match(msg) {
				m.cardIndex = 0
				m.revealed = false
				m.grade = 0
				m.score = 0
				m.input = ""
				m.cursor = 0
				m.examplesPage = 0
				return m, nil
			}
			return m, nil

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
							Grade:         ui.Grade(m.grade),
							Reverse:       m.inverse,
							UserInput:     m.input,
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
						Grade:   ui.Grade(fuzzy.Again),
						Reverse: m.inverse,
					}
				}

			case keymap.DefaultTypingQuiz.Inverse.Match(msg):
				m.inverse = !m.inverse
				m.revealed = false
				m.grade = 0
				m.score = 0
				m.input = ""
				m.cursor = 0
				m.examplesPage = 0

			default:
				m.handleInput(msg)
				return m, nil
			}

		default:
			switch {
			case keymap.DefaultTypingQuiz.Submit.Match(msg),
				keymap.DefaultList.Down.Match(msg):
				m.cardIndex++
				m.revealed = false
				m.grade = 0
				m.score = 0
				m.input = ""
				m.cursor = 0
				m.examplesPage = 0
				return m, nil

			case keymap.DefaultTypingQuiz.PrevExample.Match(msg):
				topBodyLines := strings.Count(m.renderTopBody(), "\n") + 1
				pp := quizExamplesPerPage(m.width, m.height, topBodyLines)
				if pp > 0 && m.examplesPage > 0 {
					m.examplesPage--
				}

			case keymap.DefaultTypingQuiz.NextExample.Match(msg):
				card := m.cards[m.cardIndex]
				topBodyLines := strings.Count(m.renderTopBody(), "\n") + 1
				pp := quizExamplesPerPage(m.width, m.height, topBodyLines)
				if pp > 0 {
					maxPage := (len(card.Examples) + pp - 1) / pp
					if m.examplesPage < maxPage-1 {
						m.examplesPage++
					}
				}
			}
		}
	}
	return m, nil
}

func (m *TypingQuizModel) handleInput(msg tea.KeyMsg) {
	switch msg.String() {
	case "left":
		if m.cursor > 0 {
			m.cursor--
		}
	case "right":
		runes := []rune(m.input)
		if m.cursor < len(runes) {
			m.cursor++
		}
	case "home":
		m.cursor = 0
	case "end":
		m.cursor = len([]rune(m.input))
	case "backspace":
		if m.cursor > 0 {
			runes := []rune(m.input)
			m.cursor--
			m.input = string(append(runes[:m.cursor], runes[m.cursor+1:]...))
		}
	case "delete":
		runes := []rune(m.input)
		if m.cursor < len(runes) {
			m.input = string(append(runes[:m.cursor], runes[m.cursor+1:]...))
		}
	default:
		s := msg.String()
		if s != "" && isTextInputRune(s) {
			runes := []rune(m.input)
			var b strings.Builder
			b.WriteString(string(runes[:m.cursor]))
			b.WriteString(s)
			b.WriteString(string(runes[m.cursor:]))
			m.input = b.String()
			m.cursor++
		}
	}
}

func isTextInputRune(s string) bool {
	if len(s) != 1 {
		return false
	}
	r := []rune(s)[0]
	return (r >= ' ' && r <= '~') || r > 127
}

func (m *TypingQuizModel) View() string {
	if len(m.cards) == 0 {
		return layout.Page(
			"",
			layout.Center(styles.MutedText().Render("No cards loaded"), m.width),
			componentsFooter(keymap.DefaultGlobal.Back.Help, m.width),
			m.height,
		)
	}

	if m.cardIndex >= len(m.cards) {
		return layout.Page(
			"",
			layout.Center(styles.MutedText().Render("Quiz complete!"), m.width),
			componentsFooter("enter restart · esc back", m.width),
			m.height,
		)
	}

	card := m.cards[m.cardIndex]

	var b strings.Builder

	topPad := m.height / 5
	if m.height < 10 {
		topPad = 0
	}
	b.WriteString(layout.VSpace(topPad))

	term := card.Front
	if m.inverse {
		term = strings.Join(card.Back, ", ")
	}
	b.WriteString(layout.Center(term, m.width))

	b.WriteString("\n\n")

	answerLabel := "Correct: " + m.currentCorrectAnswer()
	renderedAnswer := layout.Center(styles.MutedText().Render(answerLabel), m.width)
	answerLines := strings.Count(renderedAnswer, "\n") + 1

	if m.revealed {
		b.WriteString(renderedAnswer)
	} else {
		b.WriteString(strings.Repeat("\n", answerLines-1))
	}

	b.WriteString("\n\n")
	b.WriteString(layout.Center(m.renderInput(), m.width))

	if m.revealed {
		topBodyLines := strings.Count(m.renderTopBody(), "\n") + 1
		bottomContent := renderQuizBottomSection(card, m.width, m.height, m.examplesPage, topBodyLines)
		if bottomContent != "" {
			b.WriteString("\n\n")
			b.WriteString(bottomContent)
		}
	}

	bodyStr := b.String()
	footerStr := m.renderFooter(card)
	bodyLines := strings.Count(bodyStr, "\n") + 1
	footerLines := strings.Count(footerStr, "\n") + 1
	totalContent := bodyLines + 1 + footerLines
	if remaining := m.height - totalContent; remaining > 0 {
		b.WriteString(strings.Repeat("\n", remaining))
	}
	b.WriteString("\n\n")
	b.WriteString(footerStr)

	return b.String()
}

func (m *TypingQuizModel) renderInput() string {
	runes := []rune(m.input)
	pos := m.cursor
	if pos > len(runes) {
		pos = len(runes)
	}
	display := string(runes[:pos]) + "█" + string(runes[pos:])

	if m.revealed {
		switch m.grade {
		case fuzzy.Good:
			return ui.Theme.SuccessBg.Render(display)
		case fuzzy.Hard:
			return ui.Theme.WarningBg.Render(display)
		case fuzzy.Again:
			return ui.Theme.ErrorBg.Render(display)
		}
	}

	return ui.Theme.Background.Render(display)
}

func (m *TypingQuizModel) renderTopBody() string {
	if m.cardIndex >= len(m.cards) {
		return ""
	}
	var b strings.Builder
	topPad := m.height / 5
	if m.height < 10 {
		topPad = 0
	}
	b.WriteString(layout.VSpace(topPad))
	card := m.cards[m.cardIndex]
	term := card.Front
	if m.inverse {
		term = strings.Join(card.Back, ", ")
	}
	b.WriteString(layout.Center(term, m.width))
	b.WriteString("\n\n")
	answerLabel := "Correct: " + m.currentCorrectAnswer()
	renderedAnswer := layout.Center(styles.MutedText().Render(answerLabel), m.width)
	answerLines := strings.Count(renderedAnswer, "\n") + 1
	if m.revealed {
		b.WriteString(renderedAnswer)
	} else {
		b.WriteString(strings.Repeat("\n", answerLines-1))
	}
	b.WriteString("\n\n")
	b.WriteString(layout.Center("", m.width))
	return b.String()
}

func (m *TypingQuizModel) renderFooter(card ui.CardData) string {
	progress := fmt.Sprintf("card %d/%d", m.cardIndex+1, len(m.cards))

	if m.revealed {
		var shortcuts []string
		shortcuts = append(shortcuts, "enter next")
		shortcuts = append(shortcuts, keymap.DefaultGlobal.Back.Help)

		totalExamples := len(card.Examples)
		if totalExamples > 0 {
			topBodyLines := strings.Count(m.renderTopBody(), "\n") + 1
			pp := quizExamplesPerPage(m.width, m.height, topBodyLines)
			maxPage := (totalExamples + pp - 1) / pp
			if maxPage > 1 {
				startEx := m.examplesPage*pp + 1
				endEx := (m.examplesPage+1)*pp
				if endEx > totalExamples {
					endEx = totalExamples
				}
				shortcuts = append(shortcuts, fmt.Sprintf("ex %d-%d/%d", startEx, endEx, totalExamples))
				if m.examplesPage > 0 {
					shortcuts = append(shortcuts, keymap.DefaultTypingQuiz.PrevExample.Help)
				}
				if m.examplesPage < maxPage-1 {
					shortcuts = append(shortcuts, keymap.DefaultTypingQuiz.NextExample.Help)
				}
			}
		}

		return componentsFooter(
			progress+" · "+strings.Join(shortcuts, " · "),
			m.width,
		)
	}

	return componentsFooter(
		progress+" · "+keymap.DefaultTypingQuiz.Footer()+" · "+keymap.DefaultGlobal.Back.Help,
		m.width,
	)
}

func componentsFooter(keys string, width int) string {
	return styles.Footer(width).Render(keys)
}
