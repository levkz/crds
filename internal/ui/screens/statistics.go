package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	"crds/internal/ui/components"
	"crds/internal/ui/styles"
)

type StatisticsModel struct{}

func NewStatistics() StatisticsModel { return StatisticsModel{} }

func (m StatisticsModel) Init() tea.Cmd { return nil }

func (m StatisticsModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	return m, nil
}

func (m StatisticsModel) View() string {
	var b strings.Builder
	b.WriteString(components.Header("Statistics"))
	b.WriteString("\n\n")

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

	for _, metric := range metrics {
		b.WriteString(styles.Panel(60).Render(
			ui.Theme.Muted.Render(metric.label) + "\n" +
				ui.Theme.Primary.Render(fmt.Sprintf("%-4s", metric.value)),
		))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(components.Footer("esc back"))
	return b.String()
}
