package components

import tea "github.com/charmbracelet/bubbletea"

type NavigationKeys struct {
	Up       []string
	Down     []string
	Home     []string
	End      []string
	Confirm  []string
	Cancel   []string
	Toggle   []string
	Expand   []string
	Collapse []string
}

var DefaultNavigationKeys = NavigationKeys{
	Up:       []string{"up", "k"},
	Down:     []string{"down", "j"},
	Home:     []string{"home", "g"},
	End:      []string{"end", "G"},
	Confirm:  []string{"enter"},
	Cancel:   []string{"esc"},
	Toggle:   []string{"space"},
	Expand:   []string{"right", "l"},
	Collapse: []string{"left", "h"},
}

type TextInputKeys struct {
	Left     []string
	Right    []string
	Home     []string
	End      []string
	Back     []string
	Delete   []string
}

var DefaultTextInputKeys = TextInputKeys{
	Left:   []string{"left"},
	Right:  []string{"right"},
	Home:   []string{"home"},
	End:    []string{"end"},
	Back:   []string{"backspace"},
	Delete: []string{"delete"},
}

type CheckboxKeys struct {
	Toggle []string
}

var DefaultCheckboxKeys = CheckboxKeys{
	Toggle: []string{"space", "enter"},
}

func keyIn(msg tea.KeyMsg, keys []string) bool {
	s := msg.String()
	for _, k := range keys {
		if s == k {
			return true
		}
	}
	return false
}
