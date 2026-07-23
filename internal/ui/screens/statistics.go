package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	components "crds/internal/ui/components/display"
	"crds/internal/ui/keymap"
	"crds/internal/ui/layout"
	"crds/internal/ui/styles"
)

type StatisticsModel struct {
	stats *ui.Stats
	width  int
	height int
}

func NewStatistics() *StatisticsModel {
	return &StatisticsModel{}
}

func (m *StatisticsModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m StatisticsModel) Init() tea.Cmd { return nil }

func (m *StatisticsModel) SetStats(stats ui.Stats) {
	m.stats = &stats
}

func (m *StatisticsModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	return m, nil
}

func (m StatisticsModel) View() string {
	metrics := []struct {
		label string
		value string
	}{
		{"Reviewed Today", "0"},
		{"Accuracy", "—"},
		{"Due Today", "0"},
		{"Current Streak", "0 days"},
		{"Total Cards", "0"},
		{"Mastered", "0"},
	}

	if m.stats != nil {
		metrics[0].value = fmt.Sprintf("%d", m.stats.ReviewedToday)
		if m.stats.ReviewedToday > 0 {
			metrics[1].value = fmt.Sprintf("%.0f%%", m.stats.Accuracy)
		} else {
			metrics[1].value = "—"
		}
		metrics[4].value = fmt.Sprintf("%d", m.stats.TotalCards)
	}

	items := make([]string, len(metrics))
	for i, metric := range metrics {
		items[i] = styles.Panel(m.width).Render(
			ui.Theme.Muted.Render(metric.label) + "\n" +
				ui.Theme.Primary.Render(fmt.Sprintf("%-4s", metric.value)),
		)
	}

	return layout.Page(
		components.Header("Statistics", m.width),
		layout.Column(items...),
		components.Footer(keymap.DefaultGlobal.Back.Help, m.width),
	)
}
