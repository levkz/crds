package components

import tea "github.com/charmbracelet/bubbletea"

type SearchQueryChangedMsg struct {
	Query string
}

type SearchInputModel struct {
	TextInputModel
}

func NewSearchInput(keys ...TextInputKeys) SearchInputModel {
	k := DefaultTextInputKeys
	if len(keys) > 0 {
		k = keys[0]
	}
	return SearchInputModel{
		TextInputModel: NewTextInput(k),
	}
}

func (m SearchInputModel) Update(msg tea.Msg) (SearchInputModel, tea.Cmd) {
	prev := m.Value()
	updated, cmd := m.TextInputModel.Update(msg)
	m.TextInputModel = updated
	if m.Value() != prev {
		return m, tea.Batch(cmd, func() tea.Msg {
			return SearchQueryChangedMsg{Query: m.Value()}
		})
	}
	return m, cmd
}