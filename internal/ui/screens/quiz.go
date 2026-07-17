package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	"crds/internal/ui/components"
)

type QuizModel struct {
	CardIndex int

	Revealed bool

	Progress int

	Cards []components.Card
}

func NewQuiz() QuizModel {
	return QuizModel{}
}

func (m QuizModel) Init() tea.Cmd { return nil }

type Grade int

const (
	Again Grade = iota
	Hard
	Good
	Easy
)

func (m *QuizModel) grade(g Grade) (QuizModel, tea.Cmd) {
	return *m, nil
}

func (m QuizModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "enter":
			m.Revealed = true

		case "1":
			updated, cmd := m.grade(Again)
			return updated, cmd

		case "2":
			updated, cmd := m.grade(Hard)
			return updated, cmd

		case "3":
			updated, cmd := m.grade(Good)
			return updated, cmd

		case "4":
			updated, cmd := m.grade(Easy)
			return updated, cmd
		}
	}

	return m, nil
}

func (m QuizModel) View() string {
	var b strings.Builder

	b.WriteString(components.Header("French A1"))

	b.WriteString("\n\n")

	b.WriteString(
		components.RenderCard(
			m.Cards[m.CardIndex],
			m.Revealed,
		),
	)

	b.WriteString("\n\n")

	b.WriteString(
		components.ProgressBar(
			m.Progress,
		),
	)

	b.WriteString("\n\n")

	if m.Revealed {
		b.WriteString(
			components.Footer(
				"1 Again   2 Hard   3 Good   4 Easy",
			),
		)
	} else {
		b.WriteString(
			components.Footer(
				"Enter Reveal",
			),
		)
	}

	return b.String()
}
