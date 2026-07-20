package components

import (
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui/styles"
)

type TextInputModel struct {
	value   string
	cursor  int
	focused bool
	keys    TextInputKeys
}

func NewTextInput(keys ...TextInputKeys) TextInputModel {
	k := DefaultTextInputKeys
	if len(keys) > 0 {
		k = keys[0]
	}
	return TextInputModel{keys: k}
}

func (m TextInputModel) Init() tea.Cmd {
	return nil
}

func (m TextInputModel) Update(msg tea.Msg) (TextInputModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}
		switch {
		case keyIn(msg, m.keys.Left):
			if m.cursor > 0 {
				m.cursor--
			}
		case keyIn(msg, m.keys.Right):
			runes := []rune(m.value)
			if m.cursor < len(runes) {
				m.cursor++
			}
		case keyIn(msg, m.keys.Home):
			m.cursor = 0
		case keyIn(msg, m.keys.End):
			m.cursor = len([]rune(m.value))
		case keyIn(msg, m.keys.Back):
			if m.cursor > 0 {
				runes := []rune(m.value)
				m.cursor--
				m.value = string(append(runes[:m.cursor], runes[m.cursor+1:]...))
			}
		case keyIn(msg, m.keys.Delete):
			runes := []rune(m.value)
			if m.cursor < len(runes) {
				m.value = string(append(runes[:m.cursor], runes[m.cursor+1:]...))
			}
		default:
			s := msg.String()
			if s != "" && isTextInputRune(s) {
				runes := []rune(m.value)
				var b strings.Builder
				b.WriteString(string(runes[:m.cursor]))
				b.WriteString(s)
				b.WriteString(string(runes[m.cursor:]))
				m.value = b.String()
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m TextInputModel) View(width int) string {
	if !m.focused {
		return styles.FocusedInput().Width(width).Render(m.value)
	}
	runes := []rune(m.value)
	pos := m.cursor
	if pos > len(runes) {
		pos = len(runes)
	}
	display := string(runes[:pos]) + "█" + string(runes[pos:])
	return styles.FocusedInput().Width(width).Render(display)
}

func (m TextInputModel) Value() string {
	return m.value
}

func (m *TextInputModel) SetValue(v string) {
	m.value = v
	if m.cursor > len([]rune(v)) {
		m.cursor = len([]rune(v))
	}
}

func (m *TextInputModel) Focus()      { m.focused = true }
func (m *TextInputModel) Blur()       { m.focused = false }
func (m *TextInputModel) Focused() bool { return m.focused }
func (m *TextInputModel) Cursor() int   { return m.cursor }

func isTextInputRune(s string) bool {
	if len(s) != 1 {
		return false
	}
	r := []rune(s)[0]
	return unicode.IsPrint(r) && !unicode.IsControl(r)
}
