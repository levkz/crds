package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	components "crds/internal/ui/components/display"
	"crds/internal/ui/keymap"
	"crds/internal/ui/layout"
	"crds/internal/ui/styles"
)

type DetailModel struct {
	Term         string
	Translations []string
	Tags         []string
	Examples     []ui.ExampleData
	Notes        string
	width        int
	height       int
	examplesPage int
}

func NewDetail() *DetailModel {
	return &DetailModel{}
}

func (m *DetailModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m DetailModel) Init() tea.Cmd { return nil }

func (m *DetailModel) SetEntry(entry ui.CardData) {
	m.Term = entry.Front
	m.Translations = entry.Back
	m.Tags = entry.Tags
	m.Examples = entry.Examples
	m.Notes = entry.Notes
	m.examplesPage = 0
}

func (m *DetailModel) topPadding() int {
	if m.height < 10 {
		return 0
	}
	return m.height / 4
}

func (m *DetailModel) renderTopBody() string {
	if m.Term == "" {
		return ""
	}
	var b strings.Builder
	topPad := m.topPadding()
	if topPad > 0 {
		b.WriteString(layout.VSpace(topPad))
	}
	b.WriteString(layout.Center(m.Term, m.width))
	if len(m.Translations) > 0 {
		b.WriteString("\n\n")
		b.WriteString(layout.Center(
			styles.MutedText().Render(strings.Join(m.Translations, ", ")),
			m.width,
		))
	}
	return b.String()
}

func (m *DetailModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case keymap.DefaultQuiz.PrevExample.Match(msg):
			topBodyLines := strings.Count(m.renderTopBody(), "\n") + 1
			pp := quizExamplesPerPage(m.width, m.height, topBodyLines)
			if pp > 0 && m.examplesPage > 0 {
				m.examplesPage--
			}
		case keymap.DefaultQuiz.NextExample.Match(msg):
			topBodyLines := strings.Count(m.renderTopBody(), "\n") + 1
			pp := quizExamplesPerPage(m.width, m.height, topBodyLines)
			if pp > 0 {
				maxPage := (len(m.Examples) + pp - 1) / pp
				if m.examplesPage < maxPage-1 {
					m.examplesPage++
				}
			}
		}
	}
	return m, nil
}

func (m DetailModel) View() string {
	if m.Term == "" {
		return layout.Page(
			"",
			layout.Center(styles.MutedText().Render("Select an entry to view details"), m.width),
			components.Footer(keymap.DefaultGlobal.Back.Help, m.width),
			m.height,
		)
	}

	var b strings.Builder

	topPad := m.topPadding()
	b.WriteString(layout.VSpace(topPad))
	b.WriteString(layout.Center(m.Term, m.width))

	if len(m.Translations) > 0 {
		b.WriteString("\n\n")
		b.WriteString(layout.Center(
			styles.MutedText().Render(strings.Join(m.Translations, ", ")),
			m.width,
		))
	}

	card := ui.CardData{
		Notes:    m.Notes,
		Tags:     m.Tags,
		Examples: m.Examples,
	}
	topBodyLines := strings.Count(m.renderTopBody(), "\n") + 1
	bottomContent := renderQuizBottomSection(card, m.width, m.height, m.examplesPage, topBodyLines)
	if bottomContent != "" {
		b.WriteString("\n\n")
		b.WriteString(bottomContent)
	}

	footerStr := components.Footer(
		keymap.DefaultQuiz.PrevExample.Help+" · "+
			keymap.DefaultQuiz.NextExample.Help+" · "+
			keymap.DefaultGlobal.Back.Help,
		m.width,
	)
	bodyStr := b.String()
	bodyLines := strings.Count(bodyStr, "\n") + 1
	footerLines := strings.Count(footerStr, "\n") + 1
	if remaining := m.height - bodyLines - 1 - footerLines; remaining > 0 {
		b.WriteString(strings.Repeat("\n", remaining))
	}
	b.WriteString("\n\n")
	b.WriteString(footerStr)

	return b.String()
}
