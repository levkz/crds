package app_test

import (
	"reflect"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"crds/internal/stats"
	"crds/internal/ui"
	"crds/internal/ui/app"
	"crds/internal/ui/events"
	nav "crds/internal/ui/navigation"
)

// fakeStateScreen is a minimal ui.Screen that implements ui.StateSyncer and
// records every SyncState call.
type fakeStateScreen struct {
	called bool
	state  ui.AppState
}

func (f *fakeStateScreen) Init() tea.Cmd                       { return nil }
func (f *fakeStateScreen) Update(tea.Msg) (ui.Screen, tea.Cmd) { return f, nil }
func (f *fakeStateScreen) View() string                        { return "" }
func (f *fakeStateScreen) SetSize(w, h int)                    {}

func (f *fakeStateScreen) SyncState(s ui.AppState) tea.Cmd {
	f.called = true
	f.state = s
	return nil
}

func (f *fakeStateScreen) reset() { f.called = false }

// newTestModel returns a root Model whose Home screen is a no-op screen and
// whose given screen is a fake StateSyncer, ready to navigate to.
func newTestModel(target ui.ScreenIndex, fake *fakeStateScreen, state ui.AppState) app.Model {
	reg := nav.NewRegistry()
	reg.Register(ui.HomeScreen, &fakeStateScreen{})
	reg.Register(target, fake)
	n := nav.New(ui.HomeScreen)
	n.SetRegistry(reg)
	return app.Model{
		Config:    app.DefaultConfig(),
		State:     state,
		Navigator: n,
	}
}

func navigateTo(t *testing.T, m app.Model, target ui.ScreenIndex) app.Model {
	t.Helper()
	updated, _ := m.Update(ui.NavigateToMsg{Screen: target})
	return updated.(app.Model)
}

func TestEntrySyncOnFirstRender(t *testing.T) {
	fake := &fakeStateScreen{}
	st := ui.AppState{
		Deck: &ui.DeckData{
			Name:  "test",
			Cards: []ui.CardData{{Front: "a", Back: []string{"b"}}},
		},
	}
	m := newTestModel(ui.QuizScreen, fake, st)

	navigateTo(t, m, ui.QuizScreen)

	if !fake.called {
		t.Fatal("SyncState was not called when the screen became visible")
	}
	if !reflect.DeepEqual(fake.state, st) {
		t.Fatalf("sync state mismatch:\n got %+v\nwant %+v", fake.state, st)
	}
}

func TestStateChangeSyncsActiveScreen(t *testing.T) {
	fake := &fakeStateScreen{}
	st := ui.AppState{AllDecks: []string{"one", "two"}}
	m := newTestModel(ui.QuizScreen, fake, st)
	m = navigateTo(t, m, ui.QuizScreen)
	fake.reset()

	m.Update(ui.StateChangedMsg{State: st})

	if !fake.called {
		t.Fatal("SyncState was not called on StateChangedMsg")
	}
	if !reflect.DeepEqual(fake.state, st) {
		t.Fatalf("sync state mismatch:\n got %+v\nwant %+v", fake.state, st)
	}
}

func TestStatsLoadedMsgUpdatesStateAndSyncs(t *testing.T) {
	fake := &fakeStateScreen{}
	m := newTestModel(ui.QuizScreen, fake, ui.AppState{})
	m = navigateTo(t, m, ui.QuizScreen)
	fake.reset()

	sum := stats.Summary{TotalCards: 10}
	sel := stats.Summary{TotalCards: 4, Streak: 3}
	hist := []stats.DayPoint{{Day: "2026-08-01", Correct: 2, Incorrect: 1}}
	m2, cmd := m.Update(app.StatsLoadedMsg{Stats: sum, SelectionStats: &sel, SelectionHistory: hist})
	m3 := m2.(app.Model)

	if m3.State.Stats == nil || m3.State.Stats.TotalCards != 10 {
		t.Fatalf("State.Stats not updated: %+v", m3.State.Stats)
	}
	if m3.State.SelectionStats == nil || m3.State.SelectionStats.Streak != 3 {
		t.Fatalf("State.SelectionStats not updated: %+v", m3.State.SelectionStats)
	}
	if len(m3.State.SelectionHistory) != 1 || m3.State.SelectionHistory[0].Correct != 2 {
		t.Fatalf("State.SelectionHistory not updated: %+v", m3.State.SelectionHistory)
	}
	if cmd == nil {
		t.Fatal("expected a state-changed command from StatsLoadedMsg")
	}
	sm, ok := cmd().(ui.StateChangedMsg)
	if !ok {
		t.Fatalf("expected StateChangedMsg, got %T", cmd())
	}
	m4, _ := m3.Update(sm)
	_ = m4.(app.Model)

	if !fake.called {
		t.Fatal("SyncState was not called after StatsLoadedMsg")
	}
}

func TestTransientMessagesDoNotSync(t *testing.T) {
	fake := &fakeStateScreen{}
	m := newTestModel(ui.QuizScreen, fake, ui.AppState{})
	m = navigateTo(t, m, ui.QuizScreen)
	fake.reset()

	m.Update(ui.SaveAnswerMsg{DeckID: "d", CardID: "c", Grade: ui.GradeGood})
	if fake.called {
		t.Fatal("SyncState called for SaveAnswerMsg")
	}

	m.Update(events.TickMsg(time.Now()))
	if fake.called {
		t.Fatal("SyncState called for TickMsg")
	}

	m.Update(events.ShowNotificationMsg{Text: "hello"})
	if fake.called {
		t.Fatal("SyncState called for ShowNotificationMsg")
	}
}

func TestStateChangeSwallowedWhileOverlayActive(t *testing.T) {
	fake := &fakeStateScreen{}
	m := newTestModel(ui.QuizScreen, fake, ui.AppState{})
	m = navigateTo(t, m, ui.QuizScreen)
	fake.reset()

	m.Global.Overlay = app.ConfirmOverlay
	m.Update(ui.StateChangedMsg{State: ui.AppState{AllDecks: []string{"x"}}})

	if fake.called {
		t.Fatal("SyncState called while an overlay is active")
	}
}
