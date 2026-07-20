package components

import (
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	"crds/internal/ui/renderer"
	"crds/internal/ui/styles"
)

type SelectOptionMsg struct {
	Index int
}

type SelectModel struct {
	cursor   int
	expanded bool
	focused  bool
	keys     NavigationKeys
}

func NewSelect(keys ...NavigationKeys) SelectModel {
	k := DefaultNavigationKeys
	if len(keys) > 0 {
		k = keys[0]
	}
	return SelectModel{keys: k}
}

func (m SelectModel) Init() tea.Cmd { return nil }

func (m SelectModel) Update(msg tea.Msg) (SelectModel, tea.Cmd) {
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
		case m.expanded && keyIn(msg, m.keys.Confirm):
			m.expanded = false
			return m, func() tea.Msg { return SelectOptionMsg{Index: m.cursor} }
		case keyIn(msg, m.keys.Toggle):
			m.expanded = !m.expanded
			m.cursor = 0
		}
	}
	return m, nil
}

func (m SelectModel) View(options []string, selected int, width int) string {
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

	var b strings.Builder

	if m.expanded {
		maxItemWidth := width - 5
		if maxItemWidth < 1 {
			maxItemWidth = 1
		}
		for i, opt := range options {
			if i > 0 {
				b.WriteString("\n")
			}
			truncated := renderer.Truncate(opt, maxItemWidth)
			marker := "○"
			if i == selected {
				marker = "●"
			}
			if i == cursor {
				b.WriteString(styles.SelectedItem().Render(
					ui.Theme.Icons.Navigate + " " + marker + " " + truncated,
				))
			} else {
				b.WriteString("  " + marker + " " + truncated)
			}
		}
	} else {
		selectedLabel := ""
		if selected >= 0 && selected < len(options) {
			selectedLabel = options[selected]
		}
		display := "▼ " + selectedLabel
		s := styles.FocusedInput()
		if !m.focused {
			s = styles.MutedText()
		}
		b.WriteString(s.Width(width).Render(display))
	}

	return b.String()
}

func (m *SelectModel) Cursor() int        { return m.cursor }
func (m *SelectModel) SetCursor(n int)    { m.cursor = n }
func (m *SelectModel) Focus()             { m.focused = true }
func (m *SelectModel) Blur()              { m.focused = false }
func (m *SelectModel) Focused() bool      { return m.focused }
func (m *SelectModel) Expanded() bool     { return m.expanded }
func (m *SelectModel) Expand()            { m.expanded = true }
func (m *SelectModel) Collapse()          { m.expanded = false }
