package navigation_test

import (
	"testing"

	nav "crds/internal/ui/navigation"
)

func TestManager_ShowOverlay(t *testing.T) {
	mgr := nav.New(testHome)
	event := mgr.ShowOverlay(testQuiz)

	if !mgr.IsOverlayActive() {
		t.Error("IsOverlayActive() should be true after ShowOverlay")
	}
	if event.Overlay != testQuiz {
		t.Errorf("OverlayShownEvent.Overlay = %d, want %d", event.Overlay, testQuiz)
	}
	if event.Under != testHome {
		t.Errorf("OverlayShownEvent.Under = %d, want %d", event.Under, testHome)
	}

	// Overlay should not affect current screen
	if mgr.Current != testHome {
		t.Errorf("Current should remain %d, got %d", testHome, mgr.Current)
	}
}

func TestManager_ShowOverlayReplacesExisting(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.ShowOverlay(testQuiz)
	mgr.ShowOverlay(testSearch)

	if !mgr.IsOverlayActive() {
		t.Error("IsOverlayActive() should be true")
	}

	event, ok := mgr.HideOverlay()
	if !ok {
		t.Fatal("HideOverlay() returned ok=false")
	}
	if event.Overlay != testSearch {
		t.Errorf("OverlayHiddenEvent.Overlay = %d, want %d", event.Overlay, testSearch)
	}
	if mgr.IsOverlayActive() {
		t.Error("IsOverlayActive() should be false after HideOverlay")
	}
}

func TestManager_HideOverlay(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.ShowOverlay(testQuiz)

	event, ok := mgr.HideOverlay()
	if !ok {
		t.Fatal("HideOverlay() returned ok=false")
	}
	if mgr.IsOverlayActive() {
		t.Error("IsOverlayActive() should be false after HideOverlay")
	}
	if event.Overlay != testQuiz {
		t.Errorf("OverlayHiddenEvent.Overlay = %d, want %d", event.Overlay, testQuiz)
	}
	if event.Under != testHome {
		t.Errorf("OverlayHiddenEvent.Under = %d, want %d", event.Under, testHome)
	}

	if mgr.Current != testHome {
		t.Errorf("Current should remain %d, got %d", testHome, mgr.Current)
	}
}

func TestManager_HideOverlayEmpty(t *testing.T) {
	mgr := nav.New(testHome)
	_, ok := mgr.HideOverlay()
	if ok {
		t.Fatal("HideOverlay() with no overlay returned ok=true")
	}
}

func TestManager_OverlayDoesNotAffectStacks(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.Push(testQuiz)
	mgr.ShowOverlay(testSearch)

	if mgr.StackSize() != 1 {
		t.Errorf("StackSize = %d, want 1", mgr.StackSize())
	}
	if mgr.CanGoForward() {
		t.Error("CanGoForward() should be false (no forward navigation)")
	}
	if mgr.IsModalActive() {
		t.Error("IsModalActive() should be false")
	}

	// Dismiss overlay
	mgr.HideOverlay()

	// Back stack should be intact
	if mgr.Current != testQuiz {
		t.Errorf("Current = %d, want %d", mgr.Current, testQuiz)
	}
	if mgr.StackSize() != 1 {
		t.Errorf("StackSize = %d, want 1", mgr.StackSize())
	}

	// Regular navigation still works
	mgr.Pop()
	if mgr.Current != testHome {
		t.Errorf("Current after Pop = %d, want %d", mgr.Current, testHome)
	}
}

func TestManager_ResetClearsOverlay(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.ShowOverlay(testQuiz)

	mgr.Reset(testHome)

	if mgr.IsOverlayActive() {
		t.Error("IsOverlayActive() should be false after Reset")
	}
}

func TestManager_OverlayWithRegistry(t *testing.T) {
	r := nav.NewRegistry()
	r.Register(testHome, mockScreen{id: "home"})
	r.Register(testQuiz, mockScreen{id: "quiz"})

	mgr := nav.New(testHome)
	mgr.SetRegistry(r)

	mgr.ShowOverlay(testQuiz)

	// CurrentScreen should still return the underlying screen, not the overlay
	screen, ok := mgr.CurrentScreen()
	if !ok {
		t.Fatal("CurrentScreen() returned ok=false")
	}
	if screen.View() != "home" {
		t.Errorf("View() = %q, want %q", screen.View(), "home")
	}
}

//
// Overlay event validation
//

func TestOverlayShownEvent_Fields(t *testing.T) {
	mgr := nav.New(testHome)

	event := mgr.ShowOverlay(testQuiz)
	if event.Overlay != testQuiz {
		t.Errorf("Overlay = %d, want %d", event.Overlay, testQuiz)
	}
	if event.Under != testHome {
		t.Errorf("Under = %d, want %d", event.Under, testHome)
	}
}

func TestOverlayHiddenEvent_Fields(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.ShowOverlay(testQuiz)

	event, ok := mgr.HideOverlay()
	if !ok {
		t.Fatal("HideOverlay() returned ok=false")
	}
	if event.Overlay != testQuiz {
		t.Errorf("Overlay = %d, want %d", event.Overlay, testQuiz)
	}
	if event.Under != testHome {
		t.Errorf("Under = %d, want %d", event.Under, testHome)
	}
}

func TestOverlayHiddenEvent_Empty(t *testing.T) {
	mgr := nav.New(testHome)
	event, ok := mgr.HideOverlay()
	if ok {
		t.Fatal("expected ok=false with no overlay")
	}
	if event.Overlay != 0 || event.Under != 0 {
		t.Errorf("expected zero-value OverlayHiddenEvent, got %+v", event)
	}
}
