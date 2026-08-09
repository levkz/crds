package screens

import (
	"fmt"
	"strings"

	"crds/internal/stats"
	"crds/internal/ui"
	components "crds/internal/ui/components/display"
	"crds/internal/ui/keymap"
	"crds/internal/ui/layout"
	"crds/internal/ui/renderer"
	"crds/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type statsTab int

const (
	statsTabWords statsTab = iota
	statsTabSelection
)

type metricCell struct {
	label string
	value string
}

type StatisticsModel struct {
	selectionStats   stats.Summary
	selectionHistory []stats.DayPoint

	tab statsTab
	// per-word search state
	query        string
	cursor       int
	scrollOffset int
	results      []searchEntry
	cards        []ui.CardData
	selected     *searchEntry
	wordStats    *stats.WordStats
	wordHistory  []stats.DayPoint

	width  int
	height int
}

func NewStatistics() *StatisticsModel {
	return &StatisticsModel{tab: statsTabSelection}
}

func (m *StatisticsModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m StatisticsModel) Init() tea.Cmd { return nil }

func (m *StatisticsModel) OnEnter() tea.Cmd {
	return func() tea.Msg {
		return ui.RefreshStatsMsg{}
	}
}

func (m *StatisticsModel) OnLeave() tea.Cmd {
	m.tab = statsTabSelection
	m.query = ""
	m.results = nil
	m.cursor = 0
	m.scrollOffset = 0
	m.selected = nil
	m.wordStats = nil
	m.wordHistory = nil
	return nil
}

func (m *StatisticsModel) SyncState(s ui.AppState) tea.Cmd {
	if s.Deck != nil {
		m.cards = s.Deck.Cards
	} else {
		m.cards = nil
	}
	if s.SelectionStats != nil {
		m.selectionStats = *s.SelectionStats
	} else if s.Stats != nil {
		m.selectionStats = *s.Stats
	}
	if s.SelectionHistory != nil {
		m.selectionHistory = s.SelectionHistory
	}
	if m.query != "" {
		m.filterWordResults()
		m.cursor = 0
	}
	return nil
}

func (m *StatisticsModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case ui.WordStatsLoadedMsg:
		if m.selected != nil && msg.EntryID == m.selected.ID {
			s := msg.Stats
			m.wordStats = &s
			m.wordHistory = msg.History
		}
		return m, nil
	}
	return m, nil
}

func (m *StatisticsModel) HandleBack() bool {
	if m.tab != statsTabWords {
		return false
	}
	if len(m.query) > 0 {
		m.query = ""
		m.results = nil
		m.cursor = 0
		m.scrollOffset = 0
		return true
	}
	if m.selected != nil {
		m.selected = nil
		m.wordStats = nil
		m.wordHistory = nil
		return true
	}
	return false
}

func (m *StatisticsModel) handleKey(msg tea.KeyMsg) (ui.Screen, tea.Cmd) {
	switch {
	case keymap.DefaultStatistics.SwitchTab.Match(msg):
		m.tab = 1 - m.tab

	case keymap.DefaultList.Up.Match(msg):
		if m.tab == statsTabWords && len(m.results) > 0 && m.cursor > 0 {
			m.cursor--
			m.adjustScroll()
		}

	case keymap.DefaultList.Down.Match(msg):
		if m.tab == statsTabWords && len(m.results) > 0 && m.cursor < len(m.results)-1 {
			m.cursor++
			m.adjustScroll()
		}

	case keymap.DefaultList.Select.Match(msg):
		if m.tab == statsTabWords && len(m.results) > 0 {
			entry := m.results[m.cursor]
			m.selected = &entry
			m.wordStats = nil
			m.wordHistory = nil
			return m, func() tea.Msg {
				return ui.RefreshWordStatsMsg{EntryID: entry.ID}
			}
		}

	case keymap.DefaultSearch.DeleteChar.Match(msg):
		if m.tab == statsTabWords && len(m.query) > 0 {
			runes := []rune(m.query)
			m.query = string(runes[:len(runes)-1])
			m.filterWordResults()
			m.cursor = 0
		}

	default:
		if m.tab == statsTabWords && isPrintable(msg.String()) {
			m.query += msg.String()
			m.filterWordResults()
			m.cursor = 0
		}
	}
	return m, nil
}

func (m *StatisticsModel) filterWordResults() {
	m.results = filterCards(m.cards, m.query)
	m.scrollOffset = 0
	if m.cursor >= len(m.results) {
		m.cursor = 0
	}
}

func (m *StatisticsModel) wordMaxVisible() int {
	n := m.height - 9
	if n < 1 {
		return 1
	}
	return n
}

func (m *StatisticsModel) adjustScroll() {
	visible := m.wordMaxVisible()
	if visible <= 2 {
		m.scrollOffset = m.cursor
		return
	}
	itemLimit := visible - 2
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	if m.cursor >= m.scrollOffset+itemLimit {
		m.scrollOffset = m.cursor - itemLimit + 1
	}
}

// --- View ---

func (m StatisticsModel) View() string {
	var body, footer string
	switch m.tab {
	case statsTabSelection:
		body = m.renderSelectionTab()
		footer = keymap.DefaultStatistics.Footer() + " · " + keymap.DefaultGlobal.DeckSelect.Help + " · " + keymap.DefaultGlobal.Back.Help
	case statsTabWords:
		body = m.renderWordsTab()
		footer = keymap.DefaultStatistics.Footer() + " · " + keymap.DefaultGlobal.DeckSelect.Help + " · " + keymap.DefaultGlobal.Back.Help
		if len(m.results) > 0 {
			footer = keymap.DefaultList.Select.Help + " · " + footer
		}
	}

	return layout.Page(
		components.Header("Statistics", m.width),
		layout.Column(m.renderTabs(), body),
		components.Footer(footer, m.width),
		m.height,
	)
}

func (m StatisticsModel) renderTabs() string {
	words := ui.Theme.Muted.Render("[ Words ]")
	selection := ui.Theme.Muted.Render("[ Selection ]")
	if m.tab == statsTabWords {
		words = ui.Theme.Primary.Render("[ Words ]")
	} else {
		selection = ui.Theme.Primary.Render("[ Selection ]")
	}
	return words + " " + selection
}

func (m StatisticsModel) contentWidth() int {
	w := m.width - 4
	if w < 20 {
		return 20
	}
	return w
}

func (m StatisticsModel) renderSelectionTab() string {
	w := m.contentWidth()
	graph := components.Graph(components.ToGraphPoints(m.selectionHistory), w)
	return layout.Column(
		components.Section("Confidence over time", graph, w),
		m.renderMetricGrid(m.selectionMetrics(), w),
	)
}

func (m StatisticsModel) renderWordsTab() string {
	w := m.contentWidth()
	leftW := w * 40 / 100
	rightW := w - leftW - 2
	if leftW < 20 {
		leftW = 20
	}
	if rightW < 20 {
		rightW = 20
	}
	return lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderWordSearch(leftW),
		strings.Repeat(" ", 2),
		m.renderWordDetail(rightW),
	)
}

func (m StatisticsModel) renderWordSearch(width int) string {
	var lines []string
	lines = append(lines, ui.Theme.Header.Render(" Words "))

	input := "> " + m.query
	if m.tab == statsTabWords {
		input += "█"
	}
	lines = append(lines, renderer.Truncate(input, width))

	switch {
	case len(m.results) > 0:
		showAbove := m.scrollOffset > 0
		if showAbove {
			lines = append(lines, ui.Theme.Muted.Render(" "+ui.Theme.Icons.ArrowUp+" more "))
		}
		itemLimit := m.wordMaxVisible() - 2
		if itemLimit < 1 {
			itemLimit = 1
		}
		end := min(m.scrollOffset+itemLimit, len(m.results))
		for i := m.scrollOffset; i < end; i++ {
			r := m.results[i]
			line := r.front + " → " + strings.Join(r.back, ", ")
			truncated := renderer.Truncate(line, width)
			if i == m.cursor {
				var sb strings.Builder
				sb.WriteString(ui.Theme.Primary.Render(ui.Theme.Icons.Navigate + " "))
				for _, seg := range splitHighlight(truncated, m.query) {
					if seg.highlighted {
						sb.WriteString(ui.Theme.Secondary.Render(seg.text))
					} else {
						sb.WriteString(ui.Theme.Primary.Render(seg.text))
					}
				}
				lines = append(lines, sb.String())
			} else {
				var sb strings.Builder
				sb.WriteString(styles.MutedText().Render("  "))
				for _, seg := range splitHighlight(truncated, m.query) {
					if seg.highlighted {
						sb.WriteString(ui.Theme.Secondary.Render(seg.text))
					} else {
						sb.WriteString(styles.MutedText().Render(seg.text))
					}
				}
				lines = append(lines, sb.String())
			}
		}
		if end < len(m.results) {
			lines = append(lines, ui.Theme.Muted.Render(" "+ui.Theme.Icons.ArrowDown+" more "))
		}
	case m.query != "":
		lines = append(lines, styles.MutedText().Render("No results found"))
	default:
		lines = append(lines, styles.MutedText().Render("Type to search vocabulary"))
	}

	return strings.Join(lines, "\n")
}

func (m StatisticsModel) renderWordDetail(width int) string {
	if m.selected == nil {
		return styles.MutedText().Render("Select a word to see its statistics")
	}

	var b strings.Builder
	b.WriteString(ui.Theme.Primary.Render(m.selected.front))
	if len(m.selected.back) > 0 {
		b.WriteString(" — " + strings.Join(m.selected.back, ", "))
	}
	b.WriteString("\n\n")

	if m.wordStats == nil {
		b.WriteString(styles.MutedText().Render("Loading…"))
		return b.String()
	}

	graph := components.Graph(components.ToGraphPoints(m.wordHistory), width)
	b.WriteString(components.Section("Confidence over time", graph, width))
	b.WriteString("\n\n")
	b.WriteString(m.renderMetricGrid(m.wordMetrics(), width))
	return b.String()
}

// --- Metrics ---

func (m StatisticsModel) selectionMetrics() []metricCell {
	s := m.selectionStats
	accuracy := "—"
	if s.ReviewedToday > 0 {
		accuracy = fmt.Sprintf("%.0f%%", s.Accuracy)
	}
	due := "—"
	return []metricCell{
		{"Reviewed Today", fmt.Sprintf("%d", s.ReviewedToday)},
		{"Accuracy", accuracy},
		{"Due Today", due},
		{"Current Streak", fmt.Sprintf("%d days", s.Streak)},
		{"Total Cards", fmt.Sprintf("%d", s.TotalCards)},
		{"Mastered", fmt.Sprintf("%d", s.Mastered)},
	}
}

func (m StatisticsModel) wordMetrics() []metricCell {
	ws := m.wordStats
	if ws == nil {
		return nil
	}
	last := "—"
	if ws.LastReviewed != nil {
		last = ws.LastReviewed.Format("2006-01-02")
	}
	accuracy := "—"
	if ws.TotalReviews > 0 {
		accuracy = fmt.Sprintf("%.0f%%", ws.Accuracy())
	}
	mastered := "no"
	if ws.Mastered() {
		mastered = "yes"
	}
	return []metricCell{
		{"Reviewed Today", fmt.Sprintf("%d", ws.ReviewedToday)},
		{"Accuracy", accuracy},
		{"Due Today", "—"},
		{"Last Reviewed", last},
		{"Total Reviews", fmt.Sprintf("%d", ws.TotalReviews)},
		{"Mastered", mastered},
	}
}

func (m StatisticsModel) renderMetricGrid(cells []metricCell, width int) string {
	if len(cells) == 0 {
		return ""
	}
	cols := 3
	gap := 2
	cellW := (width - (cols-1)*gap) / cols
	if cellW < 10 {
		cellW = 10
	}

	var rows []string
	for i := 0; i < len(cells); i += cols {
		end := min(i+cols, len(cells))
		var joined []string
		for j := i; j < end; j++ {
			if j > i {
				joined = append(joined, strings.Repeat(" ", gap))
			}
			joined = append(joined, styles.Panel(cellW).Render(
				ui.Theme.Muted.Render(cells[j].label)+"\n"+
					ui.Theme.Primary.Render(cells[j].value),
			))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, joined...))
	}
	return strings.Join(rows, "\n\n")
}
