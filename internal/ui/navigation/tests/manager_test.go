package navigation_test

import (
	"testing"

	nav "crds/internal/ui/navigation"
)

func TestManager_New(t *testing.T) {
	mgr := nav.New(testHome)
	if mgr.Current != testHome {
		t.Errorf("Manager.Current = %d, want %d", mgr.Current, testHome)
	}
	if mgr.CanGoBack() {
		t.Error("new manager should not have back history")
	}
	if mgr.CanGoForward() {
		t.Error("new manager should not have forward history")
	}
	if mgr.IsModalActive() {
		t.Error("new manager should not have active modal")
	}
	if mgr.IsOverlayActive() {
		t.Error("new manager should not have active overlay")
	}
	if mgr.StackSize() != 0 {
		t.Errorf("new manager stack size = %d, want 0", mgr.StackSize())
	}
}

func TestManager_Push(t *testing.T) {
	mgr := nav.New(testHome)
	event := mgr.Push(testQuiz)

	if mgr.Current != testQuiz {
		t.Errorf("Current = %d, want %d", mgr.Current, testQuiz)
	}
	if !mgr.CanGoBack() {
		t.Error("CanGoBack() should be true after push")
	}
	if mgr.StackSize() != 1 {
		t.Errorf("StackSize() = %d, want 1", mgr.StackSize())
	}

	hist := mgr.History()
	if len(hist) != 1 || hist[0] != testHome {
		t.Errorf("History = %v, want [%d]", hist, testHome)
	}

	if event.From != testHome || event.To != testQuiz {
		t.Errorf("PushEvent = {From: %d, To: %d}, want {From: %d, To: %d}",
			event.From, event.To, testHome, testQuiz)
	}

	if mgr.CanGoForward() {
		t.Error("CanGoForward() should be false after push (forward stack cleared)")
	}
}

func TestManager_Pop(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.Push(testQuiz)
	mgr.Push(testSearch)

	event, ok := mgr.Pop()
	if !ok {
		t.Fatal("Pop() returned ok=false")
	}
	if event.To != testQuiz {
		t.Errorf("PopEvent.To = %d, want %d", event.To, testQuiz)
	}
	if event.From != testSearch {
		t.Errorf("PopEvent.From = %d, want %d", event.From, testSearch)
	}
	if mgr.Current != testQuiz {
		t.Errorf("Current after Pop = %d, want %d", mgr.Current, testQuiz)
	}
	if mgr.StackSize() != 1 {
		t.Errorf("StackSize after Pop = %d, want 1", mgr.StackSize())
	}
	if !mgr.CanGoForward() {
		t.Error("CanGoForward() should be true after pop (current saved to forward)")
	}
}

func TestManager_PopToLast(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.Push(testQuiz)

	event, ok := mgr.Pop()
	if !ok {
		t.Fatal("Pop() returned ok=false")
	}
	if event.To != testHome {
		t.Errorf("PopEvent.To = %d, want %d", event.To, testHome)
	}
	if event.From != testQuiz {
		t.Errorf("PopEvent.From = %d, want %d", event.From, testQuiz)
	}
	if mgr.Current != testHome {
		t.Errorf("Current after Pop = %d, want %d", mgr.Current, testHome)
	}
	if mgr.CanGoBack() {
		t.Error("CanGoBack() should be false after popping to initial screen")
	}
	if !mgr.CanGoForward() {
		t.Error("CanGoForward() should be true after pop (current saved to forward)")
	}
}

func TestManager_PopEmpty(t *testing.T) {
	mgr := nav.New(testHome)
	event, ok := mgr.Pop()
	if ok {
		t.Fatal("Pop() on manager with empty stack returned ok=true")
	}
	if event != (nav.PopEvent{}) {
		t.Errorf("PopEvent on empty stack should be zero value, got %+v", event)
	}
	if mgr.Current != testHome {
		t.Errorf("Current after failed Pop = %d, want %d", mgr.Current, testHome)
	}
}

func TestManager_Replace(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.Push(testQuiz)
	event := mgr.Replace(testSettings)

	if mgr.Current != testSettings {
		t.Errorf("Current after Replace = %d, want %d", mgr.Current, testSettings)
	}
	if mgr.StackSize() != 1 {
		t.Errorf("StackSize after Replace = %d, want 1", mgr.StackSize())
	}
	hist := mgr.History()
	if len(hist) != 1 || hist[0] != testHome {
		t.Errorf("History after Replace = %v, want [%d]", hist, testHome)
	}

	if event.From != testQuiz || event.To != testSettings {
		t.Errorf("ReplaceEvent = {From: %d, To: %d}, want {From: %d, To: %d}",
			event.From, event.To, testQuiz, testSettings)
	}

	popEvent, ok := mgr.Pop()
	if !ok {
		t.Fatal("Pop() after Replace returned ok=false")
	}
	if popEvent.To != testHome {
		t.Errorf("Pop().To after Replace = %d, want %d", popEvent.To, testHome)
	}
}

func TestManager_Reset(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.Push(testQuiz)
	mgr.Push(testSearch)

	event := mgr.Reset(testHome)
	if mgr.Current != testHome {
		t.Errorf("Current after Reset = %d, want %d", mgr.Current, testHome)
	}
	if mgr.CanGoBack() {
		t.Error("CanGoBack() should be false after Reset")
	}
	if mgr.CanGoForward() {
		t.Error("CanGoForward() should be false after Reset")
	}
	if mgr.IsModalActive() {
		t.Error("IsModalActive() should be false after Reset")
	}
	if mgr.IsOverlayActive() {
		t.Error("IsOverlayActive() should be false after Reset")
	}
	if mgr.StackSize() != 0 {
		t.Errorf("StackSize after Reset = %d, want 0", mgr.StackSize())
	}

	if event.To != testHome {
		t.Errorf("ResetEvent.To = %d, want %d", event.To, testHome)
	}
}

func TestManager_History(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.Push(testQuiz)
	mgr.Push(testSearch)

	hist := mgr.History()
	want := []int{int(testHome), int(testQuiz)}
	if len(hist) != len(want) {
		t.Fatalf("History length = %d, want %d", len(hist), len(want))
	}
	for i, v := range hist {
		if int(v) != want[i] {
			t.Errorf("History[%d] = %d, want %d", i, v, want[i])
		}
	}
}

func TestManager_Sequence(t *testing.T) {
	mgr := nav.New(testHome)

	mgr.Push(testQuiz)
	mgr.Push(testSearch)
	mgr.Push(testSettings)
	if mgr.Current != testSettings {
		t.Errorf("Current = %d, want %d", mgr.Current, testSettings)
	}

	mgr.Pop()
	if mgr.Current != testSearch {
		t.Errorf("Current after Pop = %d, want %d", mgr.Current, testSearch)
	}

	mgr.Replace(testSettings)
	if mgr.Current != testSettings {
		t.Errorf("Current after Replace = %d, want %d", mgr.Current, testSettings)
	}

	mgr.Pop()
	if mgr.Current != testQuiz {
		t.Errorf("Current after Pop = %d, want %d", mgr.Current, testQuiz)
	}

	mgr.Pop()
	if mgr.Current != testHome {
		t.Errorf("Current after Pop = %d, want %d", mgr.Current, testHome)
	}

	if mgr.CanGoBack() {
		t.Error("CanGoBack() should be false at end of sequence")
	}
}

//
// Manager + Registry integration tests
//

func TestManager_CurrentScreenWithoutRegistry(t *testing.T) {
	mgr := nav.New(testHome)
	_, ok := mgr.CurrentScreen()
	if ok {
		t.Error("CurrentScreen() should return ok=false when no registry is set")
	}
}

func TestManager_CurrentScreen(t *testing.T) {
	r := nav.NewRegistry()
	r.Register(testHome, mockScreen{id: "home"})
	r.Register(testQuiz, mockScreen{id: "quiz"})

	mgr := nav.New(testHome)
	mgr.SetRegistry(r)

	screen, ok := mgr.CurrentScreen()
	if !ok {
		t.Fatal("CurrentScreen() returned ok=false")
	}
	if screen.View() != "home" {
		t.Errorf("View() = %q, want %q", screen.View(), "home")
	}
}

func TestManager_CurrentScreenAfterPush(t *testing.T) {
	r := nav.NewRegistry()
	r.Register(testHome, mockScreen{id: "home"})
	r.Register(testQuiz, mockScreen{id: "quiz"})

	mgr := nav.New(testHome)
	mgr.SetRegistry(r)

	mgr.Push(testQuiz)

	screen, ok := mgr.CurrentScreen()
	if !ok {
		t.Fatal("CurrentScreen() returned ok=false")
	}
	if screen.View() != "quiz" {
		t.Errorf("View() = %q, want %q", screen.View(), "quiz")
	}
}

func TestManager_CurrentScreenAfterPop(t *testing.T) {
	r := nav.NewRegistry()
	r.Register(testHome, mockScreen{id: "home"})
	r.Register(testQuiz, mockScreen{id: "quiz"})

	mgr := nav.New(testHome)
	mgr.SetRegistry(r)
	mgr.Push(testQuiz)

	mgr.Pop()

	screen, ok := mgr.CurrentScreen()
	if !ok {
		t.Fatal("CurrentScreen() returned ok=false")
	}
	if screen.View() != "home" {
		t.Errorf("View() = %q, want %q", screen.View(), "home")
	}
}

func TestManager_CurrentScreenUnregistered(t *testing.T) {
	r := nav.NewRegistry()
	r.Register(testHome, mockScreen{id: "home"})

	mgr := nav.New(testHome)
	mgr.SetRegistry(r)
	mgr.Push(testQuiz)

	_, ok := mgr.CurrentScreen()
	if ok {
		t.Error("CurrentScreen() should return ok=false for unregistered screen")
	}
}
