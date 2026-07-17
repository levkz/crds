package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const tickInterval = 1 * time.Second

type TickMsg time.Time

func TickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
