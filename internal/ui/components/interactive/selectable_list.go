package components

import (
	"math"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	"crds/internal/ui/renderer"
	"crds/internal/ui/styles"
)

type SelectMsg struct {
	Index int
}

type SelectableListModel struct {
	cursor   int
	selected map[int]bool
	multi    bool
	focused  bool
	keys     NavigationKeys
}

func NewSelectableList(multi bool, keys ...NavigationKeys) SelectableListModel {
	k := DefaultNavigationKeys
	if len(keys) > 0 {
		k = keys[0]
	}
	return SelectableListModel{
		selected: make(map[int]bool),
		multi:    multi,
		keys:     k,
	}
}

func (m SelectableListModel) Init() tea.Cmd {
	return nil
}

func (m SelectableListModel) Update(msg tea.Msg) (SelectableListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}
		switch {
		case keyIn(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case keyIn(msg, m.keys.Down):
			m.cursor++
		case keyIn(msg, m.keys.Home):
			m.cursor = math.MaxInt
		case keyIn(msg, m.keys.End):
			m.cursor = math.MaxInt
		case keyIn(msg, m.keys.Toggle):
			if m.multi {
				m.selected[m.cursor] = !m.selected[m.cursor]
			}
		case keyIn(msg, m.keys.Confirm):
			if m.multi {
				m.selected[m.cursor] = !m.selected[m.cursor]
			}
			return m, func() tea.Msg { return SelectMsg{Index: m.cursor} }
		}
	}
	return m, nil
}

func (m SelectableListModel) View(items []string, width int) string {
	if len(items) == 0 {
		return ""
	}

	cursor := m.cursor
	if cursor >= len(items) {
		cursor = len(items) - 1
	}
	if cursor < 0 {
		cursor = 0
	}

	maxItemWidth := width - 6
	if maxItemWidth < 1 {
		maxItemWidth = 1
	}

	var b strings.Builder
	for i, item := range items {
		if i > 0 {
			b.WriteString("\n")
		}

		truncated := renderer.Truncate(item, maxItemWidth)

		selMarker := "  "
		if m.selected[i] {
			selMarker = " " + ui.Theme.Icons.Check + " "
		}

		if i == cursor && m.focused {
			line := ui.Theme.Icons.Navigate + selMarker + truncated
			b.WriteString(styles.SelectedItem().Render(line))
		} else {
			b.WriteString("  " + selMarker + truncated)
		}
	}
	return b.String()
}

func (m *SelectableListModel) clampCursor(n int) {
	if n == 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= n {
		m.cursor = n - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *SelectableListModel) Cursor() int       { return m.cursor }
func (m *SelectableListModel) SetCursor(n int)   { m.cursor = n }

func (m *SelectableListModel) Selected() []int {
	var idxs []int
	for i := range m.selected {
		if m.selected[i] {
			idxs = append(idxs, i)
		}
	}
	sort.Ints(idxs)
	return idxs
}

func (m *SelectableListModel) SelectedItems(items []string) []string {
	idxs := m.Selected()
	result := make([]string, len(idxs))
	for i, idx := range idxs {
		if idx < len(items) {
			result[i] = items[idx]
		}
	}
	return result
}

func (m *SelectableListModel) ClearSelection() {
	m.selected = make(map[int]bool)
}

func (m *SelectableListModel) Focus()      { m.focused = true }
func (m *SelectableListModel) Blur()       { m.focused = false }
func (m *SelectableListModel) Focused() bool { return m.focused }
