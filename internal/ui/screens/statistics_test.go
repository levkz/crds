package screens

import (
	"strings"
	"testing"
	"time"

	"crds/internal/stats"
	"crds/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestStatistics() *StatisticsModel {
	m := NewStatistics()
	m.SetSize(80, 24)
	return m
}

func TestStatisticsSyncState(t *testing.T) {
	m := newTestStatistics()
	sel := stats.Summary{TotalCards: 5, Streak: 2}
	hist := []stats.DayPoint{{Day: "2026-08-01", Correct: 1}}

	m.SyncState(ui.AppState{
		Deck:             &ui.DeckData{Cards: []ui.CardData{{ID: "a", Front: "a"}}},
		SelectionStats:   &sel,
		SelectionHistory: hist,
	})

	if len(m.cards) != 1 {
		t.Errorf("cards = %d, want 1", len(m.cards))
	}
	if m.selectionStats.Streak != 2 {
		t.Errorf("selectionStats.Streak = %d, want 2", m.selectionStats.Streak)
	}
	if len(m.selectionHistory) != 1 {
		t.Errorf("selectionHistory = %d, want 1", len(m.selectionHistory))
	}
}

func TestStatisticsSyncStateFallsBackToGlobalStats(t *testing.T) {
	m := newTestStatistics()
	g := stats.Summary{TotalCards: 7}
	m.SyncState(ui.AppState{Stats: &g})

	if m.selectionStats.TotalCards != 7 {
		t.Errorf("selectionStats.TotalCards = %d, want 7 (fallback)", m.selectionStats.TotalCards)
	}
}

func TestStatisticsSwitchTab(t *testing.T) {
	m := newTestStatistics()
	if m.tab != statsTabSelection {
		t.Fatalf("initial tab = %v, want selection", m.tab)
	}

	_, _ = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyTab}))
	if m.tab != statsTabWords {
		t.Errorf("tab after switch = %v, want words", m.tab)
	}

	_, _ = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyTab}))
	if m.tab != statsTabSelection {
		t.Errorf("tab after second switch = %v, want selection", m.tab)
	}
}

func TestStatisticsWordSearchFilters(t *testing.T) {
	m := newTestStatistics()
	m.tab = statsTabWords
	m.cards = []ui.CardData{
		{ID: "bonjour", Front: "bonjour", Back: []string{"hello"}},
		{ID: "merci", Front: "merci", Back: []string{"thanks"}},
	}

	for _, r := range []rune("bon") {
		_, _ = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{r}}))
	}
	if m.query != "bon" {
		t.Errorf("query = %q, want bon", m.query)
	}
	if len(m.results) != 1 || m.results[0].ID != "bonjour" {
		t.Errorf("results = %+v, want [bonjour]", m.results)
	}

	_, _ = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyBackspace}))
	_, _ = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyBackspace}))
	_, _ = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyBackspace}))
	if m.query != "" {
		t.Errorf("query after backspaces = %q, want empty", m.query)
	}
}

func TestStatisticsSelectWordEmitsRefresh(t *testing.T) {
	m := newTestStatistics()
	m.tab = statsTabWords
	m.cards = []ui.CardData{{ID: "bonjour", Front: "bonjour", Back: []string{"hello"}}}
	m.query = "bon"
	m.filterWordResults()

	_, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyEnter}))
	if cmd == nil {
		t.Fatal("expected a command from word selection")
	}
	msg := cmd()
	refresh, ok := msg.(ui.RefreshWordStatsMsg)
	if !ok {
		t.Fatalf("expected RefreshWordStatsMsg, got %T", msg)
	}
	if refresh.EntryID != "bonjour" {
		t.Errorf("EntryID = %q, want bonjour", refresh.EntryID)
	}
	if m.selected == nil || m.selected.ID != "bonjour" {
		t.Errorf("selected = %+v, want bonjour", m.selected)
	}
}

func TestStatisticsWordStatsLoaded(t *testing.T) {
	m := newTestStatistics()
	m.selected = &searchEntry{ID: "bonjour"}
	last := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	ws := stats.WordStats{TotalReviews: 3, Correct: 2, Incorrect: 1, LastReviewed: &last}
	hist := []stats.DayPoint{{Day: "2026-08-01", Correct: 2, Incorrect: 1}}

	_, _ = m.Update(ui.WordStatsLoadedMsg{EntryID: "bonjour", Stats: ws, History: hist})
	if m.wordStats == nil {
		t.Fatal("wordStats not set for matching entry")
	}
	if m.wordStats.TotalReviews != 3 {
		t.Errorf("wordStats.TotalReviews = %d, want 3", m.wordStats.TotalReviews)
	}
	if len(m.wordHistory) != 1 {
		t.Errorf("wordHistory = %d, want 1", len(m.wordHistory))
	}

	// Stale result for a different entry is ignored.
	_, _ = m.Update(ui.WordStatsLoadedMsg{EntryID: "other", Stats: stats.WordStats{TotalReviews: 9}})
	if m.wordStats.TotalReviews != 3 {
		t.Errorf("stale result overwrote wordStats: %+v", m.wordStats)
	}
}

func TestStatisticsDueTodayMetric(t *testing.T) {
	m := newTestStatistics()
	sel := stats.Summary{DueToday: 4}
	m.SyncState(ui.AppState{
		Deck:           &ui.DeckData{Cards: []ui.CardData{{ID: "a", Front: "alpha"}}},
		SelectionStats: &sel,
		Due:            []string{"a", "b"},
	})

	for _, cell := range m.selectionMetrics() {
		if cell.label == "Due Today" && cell.value != "4" {
			t.Errorf("selection Due Today = %q, want 4", cell.value)
		}
	}
}

func TestStatisticsWordDueTodayYesNo(t *testing.T) {
	m := newTestStatistics()
	m.selected = &searchEntry{ID: "a"}

	m.dueIDs = []string{"a", "c"}
	for _, cell := range m.wordMetrics() {
		if cell.label == "Due Today" && cell.value != "yes" {
			t.Errorf("due word = %q, want yes", cell.value)
		}
	}

	m.dueIDs = []string{"c"}
	for _, cell := range m.wordMetrics() {
		if cell.label == "Due Today" && cell.value != "no" {
			t.Errorf("not-due word = %q, want no", cell.value)
		}
	}
}

func TestStatisticsHandleBack(t *testing.T) {
	m := newTestStatistics()

	if m.HandleBack() {
		t.Error("HandleBack on selection tab should return false")
	}

	m.tab = statsTabWords
	m.query = "bon"
	m.cards = []ui.CardData{{ID: "bonjour", Front: "bonjour"}}
	m.filterWordResults()
	if !m.HandleBack() {
		t.Error("HandleBack with query should return true")
	}
	if m.query != "" || len(m.results) != 0 {
		t.Errorf("query/results not cleared: %q %d", m.query, len(m.results))
	}

	m.selected = &searchEntry{ID: "bonjour"}
	if !m.HandleBack() {
		t.Error("HandleBack with selection should return true")
	}
	if m.selected != nil {
		t.Error("selection not cleared")
	}

	if m.HandleBack() {
		t.Error("HandleBack with nothing to clear should return false")
	}
}

func TestStatisticsViewRenders(t *testing.T) {
	m := newTestStatistics()
	sel := stats.Summary{ReviewedToday: 3, Accuracy: 66.7, TotalCards: 5, Streak: 2, Mastered: 1}
	m.SyncState(ui.AppState{
		Deck:             &ui.DeckData{Cards: []ui.CardData{{ID: "a", Front: "alpha"}, {ID: "b", Front: "beta"}}},
		SelectionStats:   &sel,
		SelectionHistory: []stats.DayPoint{{Day: "2026-08-01", Correct: 2, Incorrect: 1}},
	})

	out := m.View()
	for _, want := range []string{"Confidence over time", "Reviewed Today", "Current Streak", "Total Cards", "Mastered"} {
		if !strings.Contains(out, want) {
			t.Errorf("selection tab missing %q", want)
		}
	}

	_, _ = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyTab}))
	out = m.View()
	for _, want := range []string{"Words", "Type to search vocabulary"} {
		if !strings.Contains(out, want) {
			t.Errorf("words tab missing %q", want)
		}
	}
}

func TestStatisticsWordDetailNoData(t *testing.T) {
	m := newTestStatistics()
	m.tab = statsTabWords
	m.selected = &searchEntry{ID: "bonjour", front: "bonjour", back: []string{"hello"}}

	out := m.renderWordDetail(40)
	if !strings.Contains(out, "Loading") {
		t.Errorf("pre-load detail missing Loading, got:\n%s", out)
	}

	_, _ = m.Update(ui.WordStatsLoadedMsg{EntryID: "bonjour", Stats: stats.WordStats{}})
	out = m.renderWordDetail(40)
	if !strings.Contains(out, "No data") {
		t.Errorf("zero-review word should show No data, got:\n%s", out)
	}

	_, _ = m.Update(ui.WordStatsLoadedMsg{EntryID: "bonjour", Stats: stats.WordStats{TotalReviews: 2}})
	out = m.renderWordDetail(40)
	if strings.Contains(out, "No data") {
		t.Errorf("reviewed word should not show No data, got:\n%s", out)
	}
	if !strings.Contains(out, "Total") {
		t.Errorf("reviewed word detail missing metrics, got:\n%s", out)
	}
}
