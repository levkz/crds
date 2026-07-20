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
	Examples     []string
	Notes        string
	width        int
	height       int
}

func NewDetail() *DetailModel {
	return &DetailModel{width: 60, height: 24}
}

func (m *DetailModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m DetailModel) Init() tea.Cmd { return nil }

func (m *DetailModel) SetEntry(entry ui.CardData) {
	m.Term = entry.Front
	m.Translations = entry.Back
	m.Notes = entry.Notes
	m.Examples = nil
}

func (m *DetailModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	return m, nil
}

func (m DetailModel) View() string {
	var sections []string

	if m.Term != "" {
		sections = append(sections, ui.Theme.Primary.Render(m.Term))
	}

	if len(m.Translations) > 0 {
		content := ui.Theme.Muted.Render("Translations") + "\n" +
			styles.Card(m.width).Render(strings.Join(m.Translations, "\n"))
		sections = append(sections, content)
	}

	if len(m.Examples) > 0 {
		content := ui.Theme.Muted.Render("Examples") + "\n" +
			styles.Card(m.width).Render(strings.Join(m.Examples, "\n"))
		sections = append(sections, content)
	}

	if m.Notes != "" {
		content := ui.Theme.Muted.Render("Notes") + "\n" +
			styles.Card(m.width).Render(m.Notes)
		sections = append(sections, content)
	}

	if m.Term == "" {
		sections = append(sections, styles.MutedText().Render("Select an entry to view details"))
	}

	return layout.Page(
		components.Header("Entry Detail", m.width),
		layout.Column(sections...),
		components.Footer(keymap.DefaultGlobal.Back.Help, m.width),
	)
}
