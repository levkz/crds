package navigation_test

import (
	"testing"

	nav "crds/internal/ui/navigation"
)

func TestManager_Forward(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.Push(testQuiz)
	mgr.Push(testSearch)

	mgr.Pop()
	if !mgr.CanGoForward() {
		t.Fatal("can't go forward after pop")
	}

	event, ok := mgr.Forward()
	if !ok {
		t.Fatal("Forward() returned ok=false")
	}
	if event.From != testQuiz {
		t.Errorf("ForwardEvent.From = %d, want %d", event.From, testQuiz)
	}
	if event.To != testSearch {
		t.Errorf("ForwardEvent.To = %d, want %d", event.To, testSearch)
	}
	if mgr.Current != testSearch {
		t.Errorf("Current after Forward = %d, want %d", mgr.Current, testSearch)
	}
	if !mgr.CanGoBack() {
		t.Error("CanGoBack() should be true after forward")
	}
}

func TestManager_ForwardEmpty(t *testing.T) {
	mgr := nav.New(testHome)
	_, ok := mgr.Forward()
	if ok {
		t.Fatal("Forward() on empty forward stack returned ok=true")
	}
	if mgr.Current != testHome {
		t.Errorf("Current after failed Forward = %d, want %d", mgr.Current, testHome)
	}
}

func TestManager_ForwardClearsBackStackNo(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.Push(testQuiz)
	mgr.Pop()

	mgr.Forward()

	if !mgr.CanGoBack() {
		t.Error("forward should push current onto back stack (CanGoBack should be true)")
	}
	if mgr.StackSize() != 1 {
		t.Errorf("StackSize after forward = %d, want 1", mgr.StackSize())
	}
}

func TestManager_PushClearsForward(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.Push(testQuiz)
	mgr.Push(testSearch)

	mgr.Pop()
	if !mgr.CanGoForward() {
		t.Fatal("forward stack should have entry after pop")
	}

	mgr.Push(testSettings)

	if mgr.CanGoForward() {
		t.Error("CanGoForward() should be false after push (forward cleared)")
	}
}

func TestManager_ResetClearsForward(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.Push(testQuiz)
	mgr.Pop()

	if !mgr.CanGoForward() {
		t.Fatal("forward should be non-empty after pop")
	}

	mgr.Reset(testHome)
	if mgr.CanGoForward() {
		t.Error("CanGoForward() should be false after reset")
	}
}

func TestManager_BackAndForwardSequence(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.Push(testQuiz)
	mgr.Push(testSearch)
	mgr.Push(testSettings)

	event, ok := mgr.Pop()
	if !ok || event.To != testSearch {
		t.Fatalf("Pop() failed or unexpected destination")
	}

	event, ok = mgr.Pop()
	if !ok || event.To != testQuiz {
		t.Fatalf("Pop() failed or unexpected destination")
	}

	if mgr.Current != testQuiz {
		t.Fatalf("Current = %d, want %d", mgr.Current, testQuiz)
	}

	fwdEvent, ok := mgr.Forward()
	if !ok {
		t.Fatal("Forward() returned ok=false")
	}
	if fwdEvent.To != testSearch {
		t.Errorf("ForwardEvent.To = %d, want %d", fwdEvent.To, testSearch)
	}
	if mgr.Current != testSearch {
		t.Errorf("Current after Forward = %d, want %d", mgr.Current, testSearch)
	}

	fwdEvent, ok = mgr.Forward()
	if !ok {
		t.Fatal("Forward() returned ok=false")
	}
	if fwdEvent.To != testSettings {
		t.Errorf("ForwardEvent.To = %d, want %d", fwdEvent.To, testSettings)
	}
	if mgr.Current != testSettings {
		t.Errorf("Current after Forward = %d, want %d", mgr.Current, testSettings)
	}

	if mgr.CanGoForward() {
		t.Error("CanGoForward() should be false at end of forward sequence")
	}

	if !mgr.CanGoBack() {
		t.Error("CanGoBack() should be true when at end of forward sequence")
	}
}

func TestManager_BackAndForwardCanGoBack(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.Push(testQuiz)
	mgr.Push(testSearch)
	mgr.Push(testSettings)

	mgr.Pop()
	mgr.Pop()
	mgr.Pop()

	if mgr.Current != testHome {
		t.Fatalf("Current after popping all = %d, want %d", mgr.Current, testHome)
	}
	if mgr.CanGoBack() {
		t.Error("CanGoBack() should be false after popping all")
	}

	if !mgr.CanGoForward() {
		t.Fatal("should be able to go forward after popping all")
	}

	mgr.Forward()
	if mgr.Current != testQuiz {
		t.Errorf("Current = %d, want %d", mgr.Current, testQuiz)
	}
	if !mgr.CanGoBack() {
		t.Error("CanGoBack() should be true after one forward")
	}

	mgr.Forward()
	if mgr.Current != testSearch {
		t.Errorf("Current = %d, want %d", mgr.Current, testSearch)
	}

	mgr.Forward()
	if mgr.Current != testSettings {
		t.Errorf("Current = %d, want %d", mgr.Current, testSettings)
	}
	if mgr.CanGoForward() {
		t.Error("CanGoForward() should be false at end")
	}
}

func TestManager_CanGoForward(t *testing.T) {
	mgr := nav.New(testHome)

	if mgr.CanGoForward() {
		t.Error("initial CanGoForward should be false")
	}

	mgr.Push(testQuiz)
	if mgr.CanGoForward() {
		t.Error("CanGoForward should be false after push")
	}

	mgr.Pop()
	if !mgr.CanGoForward() {
		t.Error("CanGoForward should be true after pop")
	}

	mgr.Forward()
	if mgr.CanGoForward() {
		t.Error("CanGoForward should be false after consuming forward")
	}
}

//
// ForwardEvent validation
//

func TestForwardEvent_Fields(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.Push(testQuiz)
	mgr.Push(testSearch)

	mgr.Pop()

	event, ok := mgr.Forward()
	if !ok {
		t.Fatal("Forward() returned ok=false")
	}
	if event.From != testQuiz {
		t.Errorf("From = %d, want %d", event.From, testQuiz)
	}
	if event.To != testSearch {
		t.Errorf("To = %d, want %d", event.To, testSearch)
	}
}

func TestForwardEvent_EmptyStack(t *testing.T) {
	mgr := nav.New(testHome)
	event, ok := mgr.Forward()
	if ok {
		t.Fatal("expected ok=false on empty forward stack")
	}
	if event.From != 0 || event.To != 0 {
		t.Errorf("expected zero-value ForwardEvent, got %+v", event)
	}
}
