package navigation_test

import (
	"testing"

	nav "crds/internal/ui/navigation"
)

func TestManager_PushModal(t *testing.T) {
	mgr := nav.New(testHome)
	event := mgr.PushModal(testQuiz)

	if mgr.Current != testQuiz {
		t.Errorf("Current after PushModal = %d, want %d", mgr.Current, testQuiz)
	}
	if !mgr.IsModalActive() {
		t.Error("IsModalActive() should be true after PushModal")
	}
	if mgr.ModalDepth() != 1 {
		t.Errorf("ModalDepth() = %d, want 1", mgr.ModalDepth())
	}
	if event.From != testHome {
		t.Errorf("ModalPushEvent.From = %d, want %d", event.From, testHome)
	}
	if event.To != testQuiz {
		t.Errorf("ModalPushEvent.To = %d, want %d", event.To, testQuiz)
	}

	if mgr.CanGoBack() {
		t.Error("CanGoBack() should still be false (no regular push happened)")
	}
}

func TestManager_DismissModal(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.PushModal(testQuiz)

	event, ok := mgr.DismissModal()
	if !ok {
		t.Fatal("DismissModal() returned ok=false")
	}
	if mgr.Current != testHome {
		t.Errorf("Current after DismissModal = %d, want %d", mgr.Current, testHome)
	}
	if mgr.IsModalActive() {
		t.Error("IsModalActive() should be false after dismissing modal")
	}
	if mgr.ModalDepth() != 0 {
		t.Errorf("ModalDepth() = %d, want 0", mgr.ModalDepth())
	}
	if event.From != testQuiz {
		t.Errorf("ModalPopEvent.From = %d, want %d", event.From, testQuiz)
	}
	if event.To != testHome {
		t.Errorf("ModalPopEvent.To = %d, want %d", event.To, testHome)
	}
}

func TestManager_DismissModalEmpty(t *testing.T) {
	mgr := nav.New(testHome)
	_, ok := mgr.DismissModal()
	if ok {
		t.Fatal("DismissModal() with no modal returned ok=true")
	}
	if mgr.Current != testHome {
		t.Errorf("Current after failed DismissModal = %d, want %d", mgr.Current, testHome)
	}
}

func TestManager_ModalDoesNotAffectBackStack(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.Push(testQuiz)
	mgr.PushModal(testSettings)

	if mgr.Current != testSettings {
		t.Errorf("Current = %d, want %d", mgr.Current, testSettings)
	}
	if !mgr.IsModalActive() {
		t.Error("IsModalActive() should be true")
	}
	if mgr.StackSize() != 1 {
		t.Errorf("StackSize = %d, want 1 (one regular push)", mgr.StackSize())
	}

	mgr.DismissModal()
	if mgr.Current != testQuiz {
		t.Errorf("Current after DismissModal = %d, want %d", mgr.Current, testQuiz)
	}
	if mgr.IsModalActive() {
		t.Error("IsModalActive() should be false after dismiss")
	}

	mgr.Pop()
	if mgr.Current != testHome {
		t.Errorf("Current after Pop = %d, want %d", mgr.Current, testHome)
	}
}

func TestManager_StackedModals(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.PushModal(testQuiz)
	mgr.PushModal(testSearch)
	mgr.PushModal(testSettings)

	if mgr.Current != testSettings {
		t.Errorf("Current = %d, want %d", mgr.Current, testSettings)
	}
	if mgr.ModalDepth() != 3 {
		t.Errorf("ModalDepth() = %d, want 3", mgr.ModalDepth())
	}

	event, ok := mgr.DismissModal()
	if !ok || event.To != testSearch {
		t.Fatalf("first DismissModal failed, To = %d", event.To)
	}
	if mgr.Current != testSearch {
		t.Errorf("Current = %d, want %d", mgr.Current, testSearch)
	}

	event, ok = mgr.DismissModal()
	if !ok || event.To != testQuiz {
		t.Fatalf("second DismissModal failed, To = %d", event.To)
	}
	if mgr.Current != testQuiz {
		t.Errorf("Current = %d, want %d", mgr.Current, testQuiz)
	}

	event, ok = mgr.DismissModal()
	if !ok || event.To != testHome {
		t.Fatalf("third DismissModal failed, To = %d", event.To)
	}
	if mgr.Current != testHome {
		t.Errorf("Current = %d, want %d", mgr.Current, testHome)
	}

	if mgr.IsModalActive() {
		t.Error("IsModalActive() should be false after dismissing all modals")
	}
}

func TestManager_ReplaceWithModal(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.PushModal(testQuiz)

	mgr.Replace(testSettings)

	if mgr.Current != testSettings {
		t.Errorf("Current after Replace = %d, want %d", mgr.Current, testSettings)
	}
	if !mgr.IsModalActive() {
		t.Error("IsModalActive() should still be true")
	}
	if mgr.ModalDepth() != 1 {
		t.Errorf("ModalDepth() = %d, want 1", mgr.ModalDepth())
	}

	mgr.DismissModal()
	if mgr.Current != testHome {
		t.Errorf("Current after DismissModal = %d, want %d", mgr.Current, testHome)
	}
}

func TestManager_ResetClearsModalStack(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.PushModal(testQuiz)
	mgr.PushModal(testSearch)

	mgr.Reset(testHome)

	if mgr.IsModalActive() {
		t.Error("IsModalActive() should be false after Reset")
	}
	if mgr.ModalDepth() != 0 {
		t.Errorf("ModalDepth() = %d, want 0", mgr.ModalDepth())
	}
}

func TestManager_CurrentScreenWithModal(t *testing.T) {
	r := nav.NewRegistry()
	r.Register(testHome, mockScreen{id: "home"})
	r.Register(testQuiz, mockScreen{id: "quiz"})

	mgr := nav.New(testHome)
	mgr.SetRegistry(r)

	mgr.PushModal(testQuiz)

	screen, ok := mgr.CurrentScreen()
	if !ok {
		t.Fatal("CurrentScreen() returned ok=false")
	}
	if screen.View() != "quiz" {
		t.Errorf("View() = %q, want %q", screen.View(), "quiz")
	}

	mgr.DismissModal()

	screen, ok = mgr.CurrentScreen()
	if !ok {
		t.Fatal("CurrentScreen() returned ok=false")
	}
	if screen.View() != "home" {
		t.Errorf("View() = %q, want %q", screen.View(), "home")
	}
}

//
// Modal event validation
//

func TestModalPushEvent_Fields(t *testing.T) {
	mgr := nav.New(testHome)

	event := mgr.PushModal(testQuiz)
	if event.From != testHome {
		t.Errorf("From = %d, want %d", event.From, testHome)
	}
	if event.To != testQuiz {
		t.Errorf("To = %d, want %d", event.To, testQuiz)
	}
}

func TestModalPopEvent_Fields(t *testing.T) {
	mgr := nav.New(testHome)
	mgr.PushModal(testQuiz)
	mgr.PushModal(testSearch)

	event, ok := mgr.DismissModal()
	if !ok {
		t.Fatal("DismissModal() returned ok=false")
	}
	if event.From != testSearch {
		t.Errorf("From = %d, want %d", event.From, testSearch)
	}
	if event.To != testQuiz {
		t.Errorf("To = %d, want %d", event.To, testQuiz)
	}
}

func TestModalPopEvent_Empty(t *testing.T) {
	mgr := nav.New(testHome)
	event, ok := mgr.DismissModal()
	if ok {
		t.Fatal("expected ok=false with no modal")
	}
	if event.From != 0 || event.To != 0 {
		t.Errorf("expected zero-value ModalPopEvent, got %+v", event)
	}
}
