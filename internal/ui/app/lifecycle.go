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

// transitionTo navigates to a screen with lifecycle hooks.
func (m Model) transitionTo(screen ui.ScreenIndex) (Model, tea.Cmd) {
	if m.Navigator.Current == screen {
		return m, nil
	}

	var cmds []tea.Cmd

	if cmd := m.leaveCurrent(); cmd != nil {
		cmds = append(cmds, cmd)
	}

	m.Navigator.Replace(screen)

	if cmd := m.enterCurrent(); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Sequence(cmds...)
}

func (m Model) popToPrevious() (Model, tea.Cmd) {
	var cmds []tea.Cmd
	if cmd := m.leaveCurrent(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	_, _ = m.Navigator.Pop()
	if cmd := m.enterCurrent(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	return m, tea.Sequence(cmds...)
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
