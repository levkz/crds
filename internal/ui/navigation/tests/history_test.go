package navigation_test

import (
	"testing"

	ui "crds/internal/ui"
	nav "crds/internal/ui/navigation"
)

//
// Stack SetLimit tests
//

func TestStack_SetLimitTrim(t *testing.T) {
	s := nav.NewStack(0)
	s.Push(testHome)
	s.Push(testQuiz)
	s.Push(testSearch)

	s.SetLimit(2)

	if s.Len() != 2 {
		t.Fatalf("Len() after SetLimit(2) = %d, want 2", s.Len())
	}

	all := s.All()
	want := []ui.ScreenIndex{testQuiz, testSearch}
	for i, v := range all {
		if v != want[i] {
			t.Errorf("All()[%d] = %d, want %d", i, v, want[i])
		}
	}
}

func TestStack_SetLimitNoTrim(t *testing.T) {
	s := nav.NewStack(0)
	s.Push(testHome)

	s.SetLimit(5)

	if s.Len() != 1 {
		t.Errorf("Len() after SetLimit(5) = %d, want 1", s.Len())
	}

	s.Push(testQuiz)
	if s.Len() != 2 {
		t.Errorf("Len() = %d, want 2", s.Len())
	}
}

func TestStack_SetLimitZero(t *testing.T) {
	s := nav.NewStack(3)
	s.Push(testHome)
	s.Push(testQuiz)
	s.Push(testSearch)

	s.SetLimit(0)

	// Since limit was 3 and already full, no trim happened
	if s.Len() != 3 {
		t.Fatalf("Len() after SetLimit(0) = %d, want 3", s.Len())
	}

	// Should now accept unlimited pushes
	s.Push(testSettings)
	if s.Len() != 4 {
		t.Errorf("Len() = %d, want 4 (unlimited after SetLimit(0))", s.Len())
	}
}

func TestStack_SetLimitRespectsPush(t *testing.T) {
	s := nav.NewStack(0)
	s.Push(testHome)
	s.Push(testQuiz)
	s.Push(testSearch)

	s.SetLimit(1)

	if s.Len() != 1 {
		t.Fatalf("Len() after SetLimit(1) = %d, want 1", s.Len())
	}
	if screen, _ := s.Peek(); screen != testSearch {
		t.Errorf("Peek() = %d, want %d (most recent retained)", screen, testSearch)
	}

	// New push should evict current
	s.Push(testSettings)
	if s.Len() != 1 {
		t.Errorf("Len() = %d, want 1", s.Len())
	}
	if screen, _ := s.Peek(); screen != testSettings {
		t.Errorf("Peek() = %d, want %d", screen, testSettings)
	}
}

//
// Manager FullHistory tests
//

func TestManager_FullHistory(t *testing.T) {
	mgr := nav.New(testHome)

	hist := mgr.FullHistory()
	want := []int{int(testHome)}
	if len(hist) != len(want) {
		t.Fatalf("FullHistory length = %d, want %d", len(hist), len(want))
	}
	if int(hist[0]) != want[0] {
		t.Errorf("FullHistory[0] = %d, want %d", hist[0], want[0])
	}
}

func TestManager_FullHistoryAfterPush(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.Push(testQuiz)
	mgr.Push(testSearch)

	hist := mgr.FullHistory()
	want := []int{int(testHome), int(testQuiz), int(testSearch)}
	if len(hist) != len(want) {
		t.Fatalf("FullHistory length = %d, want %d", len(hist), len(want))
	}
	for i, v := range hist {
		if int(v) != want[i] {
			t.Errorf("FullHistory[%d] = %d, want %d", i, v, want[i])
		}
	}

	if mgr.HistoryDepth() != 2 {
		t.Errorf("HistoryDepth() = %d, want 2", mgr.HistoryDepth())
	}
}

func TestManager_FullHistoryAfterPop(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.Push(testQuiz)
	mgr.Push(testSearch)
	mgr.Pop()

	hist := mgr.FullHistory()
	want := []int{int(testHome), int(testQuiz)}
	if len(hist) != len(want) {
		t.Fatalf("FullHistory length = %d, want %d", len(hist), len(want))
	}
	for i, v := range hist {
		if int(v) != want[i] {
			t.Errorf("FullHistory[%d] = %d, want %d", i, v, want[i])
		}
	}
}

func TestManager_FullHistoryAfterReplace(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.Push(testQuiz)
	mgr.Replace(testSearch)

	hist := mgr.FullHistory()
	want := []int{int(testHome), int(testSearch)}
	if len(hist) != len(want) {
		t.Fatalf("FullHistory length = %d, want %d", len(hist), len(want))
	}
	for i, v := range hist {
		if int(v) != want[i] {
			t.Errorf("FullHistory[%d] = %d, want %d", i, v, want[i])
		}
	}
}

//
// Manager SetMaxHistory tests
//

func TestManager_SetMaxHistoryLimitsBackStack(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.SetMaxHistory(2)

	mgr.Push(testQuiz)
	mgr.Push(testSearch)

	if mgr.StackSize() != 2 {
		t.Fatalf("StackSize = %d, want 2", mgr.StackSize())
	}

	// This push should evict testHome from history
	mgr.Push(testSettings)

	if mgr.StackSize() != 2 {
		t.Errorf("StackSize = %d, want 2", mgr.StackSize())
	}

	// testQuiz and testSearch (the screen pushed to back from current) in history, testHome evicted
	hist := mgr.History()
	want := []int{int(testQuiz), int(testSearch)}
	if len(hist) != len(want) {
		t.Fatalf("History length = %d, want %d", len(hist), len(want))
	}
	for i, v := range hist {
		if int(v) != want[i] {
			t.Errorf("History[%d] = %d, want %d", i, v, want[i])
		}
	}

	// FullHistory should include current (testSettings)
	full := mgr.FullHistory()
	fullWant := []int{int(testQuiz), int(testSearch), int(testSettings)}
	if len(full) != len(fullWant) {
		t.Fatalf("FullHistory length = %d, want %d", len(full), len(fullWant))
	}
	if mgr.Current != testSettings {
		t.Errorf("Current = %d, want %d", mgr.Current, testSettings)
	}
}

func TestManager_SetMaxHistoryUnlimited(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.SetMaxHistory(2)
	mgr.Push(testQuiz)
	mgr.Push(testSearch)

	// Set unlimited
	mgr.SetMaxHistory(0)

	// Should not evict existing entries
	if mgr.StackSize() != 2 {
		t.Fatalf("StackSize = %d, want 2", mgr.StackSize())
	}

	// Push more — should keep all
	mgr.Push(testSettings)
	if mgr.StackSize() != 3 {
		t.Errorf("StackSize = %d, want 3", mgr.StackSize())
	}
}

func TestManager_SetMaxHistoryTrimsExisting(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.SetMaxHistory(0) // unlimited first
	mgr.Push(testQuiz)
	mgr.Push(testSearch)
	mgr.Push(testSettings)

	if mgr.StackSize() != 3 {
		t.Fatalf("StackSize = %d, want 3", mgr.StackSize())
	}

	// Now set a small limit — should trim oldest
	mgr.SetMaxHistory(2)

	if mgr.StackSize() != 2 {
		t.Errorf("StackSize after SetMaxHistory(2) = %d, want 2", mgr.StackSize())
	}

	hist := mgr.History()
	want := []int{int(testQuiz), int(testSearch)}
	if len(hist) != len(want) {
		t.Fatalf("History length = %d, want %d", len(hist), len(want))
	}
	for i, v := range hist {
		if int(v) != want[i] {
			t.Errorf("History[%d] = %d, want %d", i, v, want[i])
		}
	}
}

func TestManager_HistoryDepth(t *testing.T) {
	mgr := nav.New(testHome)
	if mgr.HistoryDepth() != 0 {
		t.Errorf("HistoryDepth() = %d, want 0", mgr.HistoryDepth())
	}

	mgr.Push(testQuiz)
	if mgr.HistoryDepth() != 1 {
		t.Errorf("HistoryDepth() = %d, want 1", mgr.HistoryDepth())
	}

	mgr.Push(testSearch)
	if mgr.HistoryDepth() != 2 {
		t.Errorf("HistoryDepth() = %d, want 2", mgr.HistoryDepth())
	}

	mgr.Pop()
	if mgr.HistoryDepth() != 1 {
		t.Errorf("HistoryDepth() after Pop = %d, want 1", mgr.HistoryDepth())
	}
}
