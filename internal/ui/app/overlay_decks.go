package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"crds/internal/ui"
	components "crds/internal/ui/components/display"
	"crds/internal/ui/keymap"
)

type DeckSelectionOverlayModel struct {
	decks        []string
	selected     map[string]bool
	cursor       int
	scrollOffset int
	width        int
	height       int
}

func NewDeckSelectionOverlay() *DeckSelectionOverlayModel {
	return &DeckSelectionOverlayModel{
		selected: make(map[string]bool),
	}
}

func (m *DeckSelectionOverlayModel) SetData(decks []string, selected []string) {
	m.decks = decks
	m.selected = make(map[string]bool, len(selected))
	for _, s := range selected {
		m.selected[s] = true
	}
	if m.cursor >= len(decks) && len(decks) > 0 {
		m.cursor = len(decks) - 1
	}
	m.scrollOffset = 0
}

func (m *DeckSelectionOverlayModel) maxVisible() int {
	if m.height <= 9 {
		return 1
	}
	return m.height - 9
}

func (m *DeckSelectionOverlayModel) adjustScroll() {
	avail := m.maxVisible()
	if avail <= 2 {
		m.scrollOffset = m.cursor
		return
	}
	visible := avail - 2 // worst-case: both indicator lines shown
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	if m.cursor >= m.scrollOffset+visible {
		m.scrollOffset = m.cursor - visible + 1
	}
}

func (m *DeckSelectionOverlayModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *DeckSelectionOverlayModel) Update(msg tea.KeyMsg) tea.Cmd {
	switch {
	case keymap.DefaultDecks.Up.Match(msg):
		if m.cursor > 0 {
			m.cursor--
			m.adjustScroll()
		}

	case keymap.DefaultDecks.Down.Match(msg):
		if m.cursor < len(m.decks)-1 {
			m.cursor++
			m.adjustScroll()
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
		selected := m.selectedNames()
		return func() tea.Msg {
			return DeckOverlayConfirmMsg{Selected: selected}
		}
	}
	return nil
}

func (m *DeckSelectionOverlayModel) selectedNames() []string {
	var selected []string
	for _, name := range m.decks {
		if m.selected[name] {
			selected = append(selected, name)
		}
	}
	return selected
}

func (m *DeckSelectionOverlayModel) View() string {
	var items []string
	for _, name := range m.decks {
		prefix := " "
		if m.selected[name] {
			prefix = "✓"
		}
		items = append(items, prefix+" "+name)
	}

	modalWidth := m.width - 8
	if modalWidth < 40 {
		modalWidth = 40
	}

	var content string
	if len(m.decks) > 0 {
		content = components.RenderListClipped(items, m.cursor, m.scrollOffset, m.maxVisible(), modalWidth)
	} else {
		content = "No decks found."
	}

	footer := ui.Theme.Muted.Render(keymap.DefaultDecks.Footer() + " · " + keymap.DefaultGlobal.Back.Help)
	modal := components.RenderModal("Select Decks", content+"\n\n"+footer, modalWidth, 0)

	// Pad output to fill terminal height so fillBackground applies bg to all lines
	modalLines := strings.Count(modal, "\n") + 1
	if m.height > modalLines {
		modal += strings.Repeat("\n", m.height-modalLines)
	}
	return modal
}
