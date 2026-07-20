package components

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	"crds/internal/ui/renderer"
	"crds/internal/ui/styles"
)

type MultiSelectToggleMsg struct {
	Index int
}

type MultiSelectDoneMsg struct{}

type MultiSelectModel struct {
	cursor   int
	expanded bool
	focused  bool
	keys     NavigationKeys
}

func NewMultiSelect(keys ...NavigationKeys) MultiSelectModel {
	k := DefaultNavigationKeys
	if len(keys) > 0 {
		k = keys[0]
	}
	return MultiSelectModel{keys: k}
}

func (m MultiSelectModel) Init() tea.Cmd { return nil }

func (m MultiSelectModel) Update(msg tea.Msg) (MultiSelectModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}
		switch {
		case m.expanded && keyIn(msg, m.keys.Cancel):
			m.expanded = false
		case m.expanded && keyIn(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case m.expanded && keyIn(msg, m.keys.Down):
			m.cursor++
		case m.expanded && keyIn(msg, m.keys.Home):
			m.cursor = 0
		case m.expanded && keyIn(msg, m.keys.End):
			m.cursor = math.MaxInt
		case m.expanded && keyIn(msg, m.keys.Toggle):
			return m, func() tea.Msg { return MultiSelectToggleMsg{Index: m.cursor} }
		case m.expanded && keyIn(msg, m.keys.Confirm):
			m.expanded = false
			return m, func() tea.Msg { return MultiSelectDoneMsg{} }
		case keyIn(msg, m.keys.Toggle):
			m.expanded = !m.expanded
			m.cursor = 0
		}
	}
	return m, nil
}

func (m MultiSelectModel) View(options []string, selected map[int]bool, width int) string {
	if len(options) == 0 {
		return ""
	}

	cursor := m.cursor
	if cursor >= len(options) {
		cursor = len(options) - 1
	}
	if cursor < 0 {
		cursor = 0
	}

	if selected == nil {
		selected = make(map[int]bool)
	}

	var b strings.Builder

	if m.expanded {
		maxItemWidth := width - 6
		if maxItemWidth < 1 {
			maxItemWidth = 1
		}
		for i, opt := range options {
			if i > 0 {
				b.WriteString("\n")
			}
			truncated := renderer.Truncate(opt, maxItemWidth)
			marker := " "
			if selected[i] {
				marker = ui.Theme.Icons.Check
			}
			if i == cursor {
				b.WriteString(styles.SelectedItem().Render(
					ui.Theme.Icons.Navigate + " [" + marker + "] " + truncated,
				))
			} else {
				b.WriteString("  [" + marker + "] " + truncated)
			}
		}
	} else {
		selectedCount := 0
		for _, v := range selected {
			if v {
				selectedCount++
			}
		}
		label := selectedLabel(options, selected, selectedCount)
		display := "▼ " + label
		s := styles.FocusedInput()
		if !m.focused {
			s = styles.MutedText()
		}
		b.WriteString(s.Width(width).Render(display))
	}

	return b.String()
}

func selectedLabel(options []string, selected map[int]bool, count int) string {
	switch {
	case count == 0:
		return "None selected"
	case count == 1:
		for i, v := range selected {
			if v && i < len(options) {
				return options[i]
			}
		}
		return "?"
	default:
		first := ""
		for i, v := range selected {
			if v && i < len(options) {
				first = options[i]
				break
			}
		}
		if first == "" {
			return fmt.Sprintf("%d selected", count)
		}
		return fmt.Sprintf("%s (+%d)", first, count-1)
	}
}

func (m *MultiSelectModel) Cursor() int      { return m.cursor }
func (m *MultiSelectModel) SetCursor(n int)  { m.cursor = n }
func (m *MultiSelectModel) Focus()           { m.focused = true }
func (m *MultiSelectModel) Blur()            { m.focused = false }
func (m *MultiSelectModel) Focused() bool    { return m.focused }
func (m *MultiSelectModel) Expanded() bool   { return m.expanded }
func (m *MultiSelectModel) Expand()          { m.expanded = true }
func (m *MultiSelectModel) Collapse()        { m.expanded = false }
