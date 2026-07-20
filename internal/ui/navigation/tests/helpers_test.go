package navigation_test

import (
	tea "github.com/charmbracelet/bubbletea"
	ui "crds/internal/ui"
)

const (
	testHome     ui.ScreenIndex = iota
	testQuiz
	testSearch
	testSettings
)

type mockScreen struct {
	id string
}

func (m mockScreen) Init() tea.Cmd                        { return nil }
func (m mockScreen) Update(tea.Msg) (ui.Screen, tea.Cmd) { return m, nil }
func (m mockScreen) View() string                          { return m.id }
func (m mockScreen) SetSize(int, int)                      {}
