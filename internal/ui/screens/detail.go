package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	"crds/internal/ui/components"
	"crds/internal/ui/styles"
)

type DetailModel struct {
	Term         string
	Translations []string
	Examples     []string
	Notes        string
}

func NewDetail() DetailModel { return DetailModel{} }

func (m DetailModel) Init() tea.Cmd { return nil }

func (m DetailModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	return m, nil
}

func (m DetailModel) View() string {
	var b strings.Builder
	b.WriteString(components.Header("Entry Detail"))
	b.WriteString("\n\n")

	if m.Term != "" {
		b.WriteString(ui.Theme.Primary.Render(m.Term))
		b.WriteString("\n\n")
	}

	if len(m.Translations) > 0 {
		b.WriteString(ui.Theme.Muted.Render("Translations"))
		b.WriteString("\n")
		b.WriteString(styles.Card(60).Render(strings.Join(m.Translations, "\n")))
		b.WriteString("\n\n")
	}

	if len(m.Examples) > 0 {
		b.WriteString(ui.Theme.Muted.Render("Examples"))
		b.WriteString("\n")
		b.WriteString(styles.Card(60).Render(strings.Join(m.Examples, "\n")))
		b.WriteString("\n\n")
	}

	if m.Notes != "" {
		b.WriteString(ui.Theme.Muted.Render("Notes"))
		b.WriteString("\n")
		b.WriteString(styles.Card(60).Render(m.Notes))
		b.WriteString("\n\n")
	}

	if m.Term == "" {
		b.WriteString(styles.MutedText().Render("Select an entry to view details"))
	}

	b.WriteString("\n")
	b.WriteString(components.Footer("esc back"))
	return b.String()
}
