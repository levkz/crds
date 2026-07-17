package ui

import tea "github.com/charmbracelet/bubbletea"

type ScreenIndex int

type Screen interface {
	Init() tea.Cmd
	Update(tea.Msg) tea.Cmd
	View() string
}
