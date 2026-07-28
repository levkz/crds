package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"crds/internal/ui"
	"crds/internal/ui/keymap"
	"crds/internal/ui/layout"
	"crds/internal/ui/renderer"
	"crds/internal/ui/styles"
)

type QuizModel struct {
	cards        []ui.CardData
	cardIndex    int
	revealed     bool
	deckName     string
	inverse      bool
	examplesPage int
	width        int
	height       int
}

type Grade int

const (
	Again Grade = iota
	Hard
	Good
	Easy
)

func NewQuiz() *QuizModel {
	return &QuizModel{}
}

func (m *QuizModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *QuizModel) SetDeck(deck ui.DeckData) {
	m.cards = deck.Cards
	m.deckName = deck.Name
	m.cardIndex = 0
	m.revealed = false
	m.inverse = false
	m.examplesPage = 0
}

func (m *QuizModel) Init() tea.Cmd { return nil }

func (m *QuizModel) currentCorrectAnswer() string {
	if m.inverse {
		return m.cards[m.cardIndex].Front
	}
	return strings.Join(m.cards[m.cardIndex].Back, ", ")
}

func (m *QuizModel) grade(g Grade) (*QuizModel, tea.Cmd) {
	if m.cardIndex >= len(m.cards) {
		return m, nil
	}

	card := m.cards[m.cardIndex]
	cardID := card.Front

	m.cardIndex++
	m.revealed = false
	m.examplesPage = 0

	if m.cardIndex >= len(m.cards) {
		return m, tea.Sequence(
			func() tea.Msg {
				return ui.NavigateToMsg{Screen: ui.StatisticsScreen}
			},
		)
	}

	return m, func() tea.Msg {
		return ui.SaveAnswerMsg{
			DeckID:  m.deckName,
			CardID:  cardID,
			Grade:   int(g),
			Reverse: m.inverse,
		}
	}
}

func (m *QuizModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case keymap.DefaultGlobal.Back.Match(msg):
			return m, func() tea.Msg {
				return ui.NavigateToMsg{Screen: ui.HomeScreen}
			}

		case !m.revealed:
			switch {
			case keymap.DefaultQuiz.Reveal.Match(msg):
				m.revealed = true

			case keymap.DefaultQuiz.Inverse.Match(msg):
				m.inverse = !m.inverse
				m.revealed = false
				m.examplesPage = 0
			}

		default:
			switch {
			case keymap.DefaultQuiz.Again.Match(msg):
				return m.grade(Again)
			case keymap.DefaultQuiz.Hard.Match(msg):
				return m.grade(Hard)
			case keymap.DefaultQuiz.Good.Match(msg):
				return m.grade(Good)
			case keymap.DefaultQuiz.Easy.Match(msg):
				return m.grade(Easy)

			case keymap.DefaultQuiz.PrevExample.Match(msg):
				card := m.cards[m.cardIndex]
				pp := m.examplesPerPage(card.Examples)
				if pp > 0 && m.examplesPage > 0 {
					m.examplesPage--
				}

			case keymap.DefaultQuiz.NextExample.Match(msg):
				card := m.cards[m.cardIndex]
				pp := m.examplesPerPage(card.Examples)
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

func (m *QuizModel) View() string {
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
			componentsFooter(keymap.DefaultGlobal.Back.Help, m.width),
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
		bottomContent := m.renderBottomSection(card)
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
	var highlight lipgloss.Style = ui.Theme.Primary

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

func (m *QuizModel) renderBottomSection(card ui.CardData) string {
	sidePad := lipgloss.NewStyle().PaddingLeft(8).PaddingRight(8)
	var parts []string

	if card.Notes != "" {
		parts = append(parts, sidePad.Render(styles.MutedText().Render("note: "+card.Notes)))
	}

	if len(card.Tags) > 0 {
		parts = append(parts, sidePad.Render(m.renderTags(card.Tags)))
	}

	if len(card.Examples) > 0 {
		parts = append(parts, sidePad.Render(m.renderExamplesBlock(card.Examples)))
	}

	return strings.Join(parts, "\n\n")
}

func (m *QuizModel) renderTags(tags []string) string {
	var styled []string
	tagStyle := styles.PrimaryBg().Padding(0, 1)
	for _, t := range tags {
		styled = append(styled, tagStyle.Render(t))
	}
	return strings.Join(styled, " ")
}

func (m *QuizModel) renderExamplesBlock(examples []ui.ExampleData) string {
	pp := m.examplesPerPage(examples)
	if pp <= 0 {
		pp = 1
	}
	start := m.examplesPage * pp
	if start >= len(examples) {
		return ""
	}
	end := start + pp
	if end > len(examples) {
		end = len(examples)
	}
	page := examples[start:end]

	if m.width > 80 {
		return m.renderExamplesTwoCol(page)
	}
	return m.renderExamplesSingleCol(page)
}

func (m *QuizModel) renderExamplesSingleCol(examples []ui.ExampleData) string {
	colWidth := m.width - 2
	if colWidth < 10 {
		colWidth = 10
	}
	var blocks []string
	for _, ex := range examples {
		var blockLines []string
		blockLines = append(blockLines, renderer.Wrap("- "+ex.Text, colWidth)...)
		if ex.Translation != "" {
			blockLines = append(blockLines, renderer.Wrap("  "+ex.Translation, colWidth)...)
		}
		blocks = append(blocks, strings.Join(blockLines, "\n"))
	}
	return strings.Join(blocks, "\n\n")
}

func (m *QuizModel) renderExamplesTwoCol(examples []ui.ExampleData) string {
	colWidth := (m.width - 3) / 2
	if colWidth < 10 {
		colWidth = 10
	}

	var rows []string
	for i := 0; i < len(examples); i += 2 {
		left := m.renderExampleCell(examples[i], colWidth)
		var right []string
		if i+1 < len(examples) {
			right = m.renderExampleCell(examples[i+1], colWidth)
		}
		maxH := len(left)
		if len(right) > maxH {
			maxH = len(right)
		}
		for len(left) < maxH {
			left = append(left, strings.Repeat(" ", colWidth))
		}
		for len(right) < maxH {
			right = append(right, strings.Repeat(" ", colWidth))
		}
		for j := 0; j < maxH; j++ {
			rows = append(rows, left[j]+" "+right[j])
		}
		rows = append(rows, "")
	}

	return strings.Join(rows, "\n")
}

func (m *QuizModel) renderExampleCell(ex ui.ExampleData, width int) []string {
	var lines []string
	lines = append(lines, renderer.Wrap("- "+ex.Text, width)...)
	if ex.Translation != "" {
		lines = append(lines, renderer.Wrap("  "+ex.Translation, width)...)
	}
	for i, l := range lines {
		if w := renderer.VisibleWidth(l); w < width {
			lines[i] = l + strings.Repeat(" ", width-w)
		}
	}
	return lines
}

func (m *QuizModel) examplesPerPage(examples []ui.ExampleData) int {
	if len(examples) == 0 {
		return 0
	}
	bodyStr := m.renderTopBody()
	topLines := strings.Count(bodyStr, "\n") + 1
	availLines := m.height - topLines - 3

	if availLines < 3 {
		return 1
	}

	if m.width > 80 {
		perRow := 2
		linesPerItem := 2
		itemsPerPage := (availLines / (linesPerItem + 1)) * perRow
		if itemsPerPage < perRow {
			itemsPerPage = perRow
		}
		return itemsPerPage
	}
	linesPerItem := 3
	itemsPerPage := availLines / (linesPerItem + 1)
	if itemsPerPage < 1 {
		itemsPerPage = 1
	}
	return itemsPerPage
}

func (m *QuizModel) renderTopBody() string {
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
	progress := fmt.Sprintf("card %d/%d", m.cardIndex+1, len(m.cards))

	if m.revealed {
		var shortcuts []string
		shortcuts = append(shortcuts, keymap.DefaultQuiz.Revealed())
		shortcuts = append(shortcuts, keymap.DefaultGlobal.Back.Help)

		totalExamples := len(card.Examples)
		if totalExamples > 0 {
			pp := m.examplesPerPage(card.Examples)
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
