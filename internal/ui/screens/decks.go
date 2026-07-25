package screens

import (
	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	components "crds/internal/ui/components/display"
	"crds/internal/ui/keymap"
	"crds/internal/ui/layout"
)

type DecksModel struct {
	decks    []string
	selected map[string]bool
	cursor   int
	width    int
	height   int
}

func NewDecks() *DecksModel {
	return &DecksModel{
		selected: make(map[string]bool),
	}
}

func (m *DecksModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *DecksModel) SetDecks(decks []string, selected []string) {
	m.decks = decks
	m.selected = make(map[string]bool, len(selected))
	for _, s := range selected {
		m.selected[s] = true
	}
	if m.cursor >= len(decks) && len(decks) > 0 {
		m.cursor = len(decks) - 1
	}
}

func (m *DecksModel) Init() tea.Cmd { return nil }

func (m *DecksModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case keymap.DefaultGlobal.Back.Match(msg):
			return m, func() tea.Msg {
				return ui.NavigateToMsg{Screen: ui.HomeScreen}
			}

		case keymap.DefaultDecks.Up.Match(msg):
			if m.cursor > 0 {
				m.cursor--
			}

		case keymap.DefaultDecks.Down.Match(msg):
			if m.cursor < len(m.decks)-1 {
				m.cursor++
			}

		case keymap.DefaultDecks.Toggle.Match(msg):
			if len(m.decks) > 0 {
				name := m.decks[m.cursor]
				m.selected[name] = !m.selected[name]
			}

		case keymap.DefaultDecks.ToggleAll.Match(msg):
			all := len(m.decks) > 0
			for _, name := range m.decks {
				if !m.selected[name] {
					all = false
					break
				}
			}
			for _, name := range m.decks {
				m.selected[name] = !all
			}

		case keymap.DefaultDecks.Select.Match(msg):
			var selected []string
			for _, name := range m.decks {
				if m.selected[name] {
					selected = append(selected, name)
				}
			}
			return m, tea.Sequence(
				func() tea.Msg {
					return ui.DeckSelectionChangedMsg{Selected: selected}
				},
				func() tea.Msg {
					return ui.NavigateToMsg{Screen: ui.HomeScreen}
				},
			)
		}
	}
	return m, nil
}

func (m *DecksModel) View() string {
	var items []string
	for _, name := range m.decks {
		prefix := " "
		if m.selected[name] {
			prefix = "✓"
		}
		items = append(items, prefix+" "+name)
	}

	content := components.RenderList(items, m.cursor, m.width)
	if len(m.decks) == 0 {
		content = "No decks found."
	}

	return layout.Page(
		components.Header("Decks", m.width),
		content,
		components.Footer(keymap.DefaultDecks.Footer()+" · "+keymap.DefaultGlobal.Back.Help, m.width),
		m.height,
	)
}
