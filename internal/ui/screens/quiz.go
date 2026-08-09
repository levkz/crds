package screens

import (
	"fmt"
	"reflect"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"crds/internal/stats"
	"crds/internal/ui"
	"crds/internal/ui/keymap"
	"crds/internal/ui/layout"
	"crds/internal/ui/styles"
)

type QuizModel struct {
	cards         []ui.CardData
	originalCards []ui.CardData
	cardProgress  map[string]stats.EntryProgress
	dueIDs        []string
	mode          ui.QuizMode
	cardIndex     int
	revealed      bool
	inProgress    bool
	deckName      string
	inverse       bool
	examplesPage  int
	width         int
	height        int
}

func NewQuiz() *QuizModel {
	return &QuizModel{}
}

func (m *QuizModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *QuizModel) SyncState(s ui.AppState) tea.Cmd {
	changed := false
	if s.Deck != nil {
		if m.deckName != s.Deck.Name || !reflect.DeepEqual(m.originalCards, s.Deck.Cards) {
			m.originalCards = s.Deck.Cards
			m.deckName = s.Deck.Name
			m.inverse = false
			m.examplesPage = 0
			m.inProgress = false
			changed = true
		}
	}
	if !reflect.DeepEqual(m.cardProgress, s.DeckProgress) {
		m.cardProgress = s.DeckProgress
		changed = true
	}
	if !reflect.DeepEqual(m.dueIDs, s.Due) {
		m.dueIDs = s.Due
		changed = true
	}
	if m.mode != s.QuizMode {
		m.mode = s.QuizMode
		changed = true
	}
	// Session snapshot: once the quiz is in progress, progress/due refreshes
	// update the stored data but never reshuffle the remaining cards.
	if changed && !m.inProgress {
		m.applySort()
	}
	return nil
}

func (m *QuizModel) SetModeFromKey() tea.Cmd {
	m.mode = m.mode.Next()
	m.applySort()
	return func() tea.Msg {
		return ui.SetQuizModeMsg{Mode: m.mode}
	}
}

func (m *QuizModel) applySort() {
	m.cards = ui.SortCards(m.mode, m.originalCards, m.cardProgress, m.dueIDs)
	m.cardIndex = 0
	m.revealed = false
	m.inProgress = false
}

func (m *QuizModel) Init() tea.Cmd { return nil }

func (m *QuizModel) IsInProgress() bool {
	return m.cardIndex > 0 && m.cardIndex < len(m.cards)
}

func (m *QuizModel) currentCorrectAnswer() string {
	if m.inverse {
		return m.cards[m.cardIndex].Front
	}
	return strings.Join(m.cards[m.cardIndex].Back, ", ")
}

func (m *QuizModel) grade(g ui.Grade) (*QuizModel, tea.Cmd) {
	if m.cardIndex >= len(m.cards) {
		return m, nil
	}

	card := m.cards[m.cardIndex]

	m.cardIndex++
	m.revealed = false
	m.examplesPage = 0
	m.inProgress = true

	return m, func() tea.Msg {
		return ui.SaveAnswerMsg{
			DeckID:  card.DeckID,
			CardID:  card.ID,
			Grade:   g,
			Reverse: m.inverse,
		}
	}
}

func (m *QuizModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case keymap.DefaultGlobal.Back.Match(msg):
			if len(m.cards) > 0 && m.cardIndex >= len(m.cards) {
				return m, func() tea.Msg {
					return ui.NavigateToMsg{Screen: ui.HomeScreen}
				}
			}
			return m, func() tea.Msg {
				return ui.NavigateToMsg{Screen: ui.HomeScreen}
			}

		case len(m.cards) > 0 && m.cardIndex >= len(m.cards):
			if keymap.DefaultQuiz.Reveal.Match(msg) {
				m.applySort()
				return m, nil
			}
			return m, nil

		case !m.revealed:
			switch {
			case keymap.DefaultQuiz.Reveal.Match(msg):
				m.revealed = true

			case keymap.DefaultQuiz.Inverse.Match(msg):
				m.inverse = !m.inverse
				m.revealed = false
				m.examplesPage = 0

			case keymap.DefaultQuiz.ModeCycle.Match(msg):
				return m, m.SetModeFromKey()
			}

		default:
			switch {
			case keymap.DefaultQuiz.Again.Match(msg):
				return m.grade(ui.GradeAgain)
			case keymap.DefaultQuiz.Hard.Match(msg):
				return m.grade(ui.GradeHard)
			case keymap.DefaultQuiz.Good.Match(msg):
				return m.grade(ui.GradeGood)
			case keymap.DefaultQuiz.Easy.Match(msg):
				return m.grade(ui.GradeEasy)

			case keymap.DefaultQuiz.PrevExample.Match(msg):
				topBodyLines := strings.Count(m.renderTopBody(), "\n") + 1
				pp := quizExamplesPerPage(m.width, m.height, topBodyLines)
				if pp > 0 && m.examplesPage > 0 {
					m.examplesPage--
				}

			case keymap.DefaultQuiz.NextExample.Match(msg):
				card := m.cards[m.cardIndex]
				topBodyLines := strings.Count(m.renderTopBody(), "\n") + 1
				pp := quizExamplesPerPage(m.width, m.height, topBodyLines)
				if pp > 0 {
					maxPage := (len(card.Examples) + pp - 1) / pp
					if m.examplesPage < maxPage-1 {
						m.examplesPage++
					}
				}

			case keymap.DefaultQuiz.ModeCycle.Match(msg):
				return m, m.SetModeFromKey()
			}
		}
	}
	return m, nil
}

func (m *QuizModel) View() string {
	if len(m.cards) == 0 {
		msg := "No cards loaded"
		if m.mode == ui.QuizModeDue && len(m.originalCards) > 0 {
			msg = "No cards due"
		}
		return layout.Page(
			"",
			layout.Center(styles.MutedText().Render(msg), m.width),
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

	topPad := m.height / 4
	if m.height < 10 {
		topPad = 0
	}
	b.WriteString(layout.VSpace(topPad))

	term := card.Front
	if m.inverse {
		term = strings.Join(card.Back, ", ")
	}
	b.WriteString(layout.Center(term, m.width))

	if m.revealed {
		b.WriteString("\n\n")
		b.WriteString(layout.Center(
			styles.MutedText().Render("Correct: "+m.currentCorrectAnswer()),
			m.width,
		))

		b.WriteString("\n\n")
		b.WriteString(m.renderGradeMenu())
	}

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

func (m *QuizModel) renderGradeMenu() string {
	type gradeItem struct {
		name    string
		primary string
	}

	items := []gradeItem{
		{"again", keymap.DefaultQuiz.Again.Keys[0]},
		{"hard", keymap.DefaultQuiz.Hard.Keys[0]},
		{"okay", keymap.DefaultQuiz.Good.Keys[0]},
		{"easy", keymap.DefaultQuiz.Easy.Keys[0]},
	}

	var cells []string
	for i, item := range items {
		cells = append(cells, renderMenuLetterItem(item.name, item.primary))
		if i < len(items)-1 {
			cells = append(cells, "  ")
		}
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
	return layout.Center(row, m.width)
}

func renderMenuLetterItem(name, primary string) string {
	highlight := ui.Theme.Primary

	if isSingleLetter(primary) {
		lowerPrimary := strings.ToLower(primary)
		if idx := strings.Index(strings.ToLower(name), lowerPrimary); idx >= 0 {
			before := name[:idx]
			letter := string(name[idx])
			if primary != strings.ToLower(primary) {
				letter = primary
			}
			after := name[idx+len(lowerPrimary):]
			return ui.Theme.Muted.Render(before) +
				highlight.Render("["+letter+"]") +
				ui.Theme.Muted.Render(after)
		}
	}

	display := shortcutDisplay(primary)
	return ui.Theme.Muted.Render(name) + " " + highlight.Render("["+display+"]")
}

func (m *QuizModel) renderTopBody() string {
	if m.cardIndex >= len(m.cards) {
		return ""
	}
	var b strings.Builder
	topPad := m.height / 4
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
	if m.revealed {
		b.WriteString("\n\n")
		b.WriteString(layout.Center(
			styles.MutedText().Render("Correct: "+m.currentCorrectAnswer()),
			m.width,
		))
		b.WriteString("\n\n")
		b.WriteString(m.renderGradeMenu())
	}
	b.WriteString("\n\n")
	b.WriteString(layout.Center("", m.width))
	return b.String()
}

func (m *QuizModel) renderFooter(card ui.CardData) string {
	progress := fmt.Sprintf("card %d/%d · %s", m.cardIndex+1, len(m.cards), m.mode)

	if m.revealed {
		var shortcuts []string
		shortcuts = append(shortcuts, keymap.DefaultQuiz.Revealed())
		shortcuts = append(shortcuts, keymap.DefaultGlobal.Back.Help)

		totalExamples := len(card.Examples)
		if totalExamples > 0 {
			topBodyLines := strings.Count(m.renderTopBody(), "\n") + 1
			pp := quizExamplesPerPage(m.width, m.height, topBodyLines)
			maxPage := (totalExamples + pp - 1) / pp
			if maxPage > 1 {
				startEx := m.examplesPage*pp + 1
				endEx := (m.examplesPage + 1) * pp
				if endEx > totalExamples {
					endEx = totalExamples
				}
				shortcuts = append(shortcuts, fmt.Sprintf("ex %d-%d/%d", startEx, endEx, totalExamples))
				if m.examplesPage > 0 {
					shortcuts = append(shortcuts, keymap.DefaultQuiz.PrevExample.Help)
				}
				if m.examplesPage < maxPage-1 {
					shortcuts = append(shortcuts, keymap.DefaultQuiz.NextExample.Help)
				}
			}
		}

		return componentsFooter(
			progress+" · "+strings.Join(shortcuts, " · "),
			m.width,
		)
	}

	return componentsFooter(
		progress+" · "+keymap.DefaultQuiz.Unrevealed()+" · "+keymap.DefaultGlobal.Back.Help,
		m.width,
	)
}
