package components

import (
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	"crds/internal/ui/renderer"
	"crds/internal/ui/styles"
)

type RadioSelectedMsg struct {
	Index int
}

type RadioGroupModel struct {
	cursor  int
	focused bool
	keys    NavigationKeys
}

func NewRadioGroup(keys ...NavigationKeys) RadioGroupModel {
	k := DefaultNavigationKeys
	if len(keys) > 0 {
		k = keys[0]
	}
	return RadioGroupModel{keys: k}
}

func (m RadioGroupModel) Init() tea.Cmd { return nil }

func (m RadioGroupModel) Update(msg tea.Msg) (RadioGroupModel, tea.Cmd) {
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
			m.cursor = 0
		case keyIn(msg, m.keys.End):
			m.cursor = math.MaxInt
		case keyIn(msg, m.keys.Confirm):
			return m, func() tea.Msg { return RadioSelectedMsg{Index: m.cursor} }
		}
	}
	return m, nil
}

func (m RadioGroupModel) View(options []string, selected int, width int) string {
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

	maxItemWidth := width - 5
	if maxItemWidth < 1 {
		maxItemWidth = 1
	}

	var b strings.Builder
	for i, opt := range options {
		if i > 0 {
			b.WriteString("\n")
		}

		truncated := renderer.Truncate(opt, maxItemWidth)

		marker := "○"
		if i == selected {
			marker = "●"
		}

		if i == cursor && m.focused {
			b.WriteString(styles.SelectedItem().Render(
				ui.Theme.Icons.Navigate + " " + marker + " " + truncated,
			))
		} else {
			b.WriteString("  " + marker + " " + truncated)
		}
	}
	return b.String()
}

func (m *RadioGroupModel) Cursor() int     { return m.cursor }
func (m *RadioGroupModel) SetCursor(n int) { m.cursor = n }
func (m *RadioGroupModel) Focus()          { m.focused = true }
func (m *RadioGroupModel) Blur()           { m.focused = false }
func (m *RadioGroupModel) Focused() bool   { return m.focused }
