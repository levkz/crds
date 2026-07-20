package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	"crds/internal/ui/styles"
)

type CheckboxToggleMsg struct{}

type CheckboxModel struct {
	focused bool
	keys    CheckboxKeys
}

func NewCheckbox(keys ...CheckboxKeys) CheckboxModel {
	k := DefaultCheckboxKeys
	if len(keys) > 0 {
		k = keys[0]
	}
	return CheckboxModel{keys: k}
}

func (m CheckboxModel) Init() tea.Cmd {
	return nil
}

func (m CheckboxModel) Update(msg tea.Msg) (CheckboxModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if keyIn(msg, m.keys.Toggle) && m.focused {
			return m, func() tea.Msg { return CheckboxToggleMsg{} }
		}
	}
	return m, nil
}

func (m CheckboxModel) View(checked bool, label string, width int) string {
	style := styles.MutedText()
	if m.focused {
		style = styles.SelectedItem()
	}

	marker := " "
	if checked {
		marker = ui.Theme.Icons.Check
	}

	display := "[" + marker + "] " + label
	return style.Width(width).Render(display)
}

func (m *CheckboxModel) Focus()       { m.focused = true }
func (m *CheckboxModel) Blur()        { m.focused = false }
func (m *CheckboxModel) Focused() bool  { return m.focused }
