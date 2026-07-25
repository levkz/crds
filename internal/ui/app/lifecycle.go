package app

import (
	tea "github.com/charmbracelet/bubbletea"
	ui "crds/internal/ui"
)

// ShutdownMsg is emitted when shutdown cleanup is complete
type ShutdownMsg struct{}

// ShutdownCmd returns a command that performs graceful cleanup and signals completion
func (m Model) ShutdownCmd() tea.Cmd {
	return func() tea.Msg {
		return ShutdownMsg{}
	}
}

// Lifecycle defines optional hooks for screen enter/leave transitions.
type Lifecycle interface {
	OnEnter() tea.Cmd
	OnLeave() tea.Cmd
}

// transitionTo navigates to a screen using Replace (flat navigation, no history).
func (m Model) transitionTo(screen ui.ScreenIndex) (Model, tea.Cmd) {
	if m.Navigator.Current == screen {
		return m, nil
	}

	var cmds []tea.Cmd

	if cmd := m.leaveCurrent(); cmd != nil {
		cmds = append(cmds, cmd)
	}

	m.Navigator.Replace(screen)

	if s, ok := m.Navigator.CurrentScreen(); ok {
		s.SetSize(m.Width, m.Height)
	}

	if cmd := m.enterCurrent(); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Sequence(cmds...)
}

// pushTo navigates to a screen using Push (stacked navigation, enables back).
// The current screen is preserved in the history stack — no OnLeave is called.
func (m Model) pushTo(screen ui.ScreenIndex) (Model, tea.Cmd) {
	if m.Navigator.Current == screen {
		return m, nil
	}

	m.Navigator.Push(screen)

	if s, ok := m.Navigator.CurrentScreen(); ok {
		s.SetSize(m.Width, m.Height)
	}

	if cmd := m.enterCurrent(); cmd != nil {
		return m, cmd
	}

	return m, nil
}

// popToPrevious navigates back using Pop. The current screen is pushed to the
// forward stack — no OnLeave is called. OnEnter fires on the restored screen.
func (m Model) popToPrevious() (Model, tea.Cmd) {
	_, _ = m.Navigator.Pop()
	if s, ok := m.Navigator.CurrentScreen(); ok {
		s.SetSize(m.Width, m.Height)
	}
	if cmd := m.enterCurrent(); cmd != nil {
		return m, cmd
	}
	return m, nil
}

func (m Model) leaveCurrent() tea.Cmd {
	screen, ok := m.Navigator.CurrentScreen()
	if !ok {
		return nil
	}
	if l, ok := screen.(Lifecycle); ok {
		return l.OnLeave()
	}
	return nil
}

func (m Model) enterCurrent() tea.Cmd {
	screen, ok := m.Navigator.CurrentScreen()
	if !ok {
		return nil
	}
	if l, ok := screen.(Lifecycle); ok {
		return l.OnEnter()
	}
	return nil
}
