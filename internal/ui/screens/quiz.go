package screens

import (
	"crds/internal/ui/components"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type QuizModel struct {
	CardIndex int

	Revealed bool

	Progress int

	Cards []Card
}

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

func (m QuizModel) Update(msg tea.Msg) (QuizModel, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "enter":
			m.Revealed = true

		case "1":
			return m.grade(Again)

		case "2":
			return m.grade(Hard)

		case "3":
			return m.grade(Good)

		case "4":
			return m.grade(Easy)
		}
	}

	return m, nil
}

func (m QuizModel) View() string {

	var b strings.Builder

	b.WriteString(components.Header("French A1"))

	b.WriteString("\n\n")

	b.WriteString(
		RenderCard(
			m.Cards[m.CardIndex],
			m.Revealed,
		),
	)

	b.WriteString("\n\n")

	b.WriteString(
		ProgressBar(
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
