package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui/events"
)

const tickInterval = 1 * time.Second

func TickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return events.TickMsg(t)
	})
}
