package navigation_test

import (
	"testing"

	ui "crds/internal/ui"
	nav "crds/internal/ui/navigation"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := nav.NewRegistry()
	r.Register(testHome, mockScreen{id: "home"})

	screen, ok := r.Get(testHome)
	if !ok {
		t.Fatal("Get() returned ok=false")
	}
	if screen.View() != "home" {
		t.Errorf("View() = %q, want %q", screen.View(), "home")
	}
}

func TestRegistry_GetMissing(t *testing.T) {
	r := nav.NewRegistry()
	_, ok := r.Get(ui.ScreenIndex(99))
	if ok {
		t.Fatal("Get() for unregistered index returned ok=true")
	}
}

func TestRegistry_Has(t *testing.T) {
	r := nav.NewRegistry()
	if r.Has(testHome) {
		t.Error("Has() should be false before registration")
	}

	r.Register(testHome, mockScreen{id: "home"})
	if !r.Has(testHome) {
		t.Error("Has() should be true after registration")
	}

	if r.Has(ui.ScreenIndex(99)) {
		t.Error("Has() should be false for unregistered index")
	}
}

func TestRegistry_Remove(t *testing.T) {
	r := nav.NewRegistry()
	r.Register(testHome, mockScreen{id: "home"})
	if !r.Has(testHome) {
		t.Fatal("registration failed")
	}

	r.Remove(testHome)
	if r.Has(testHome) {
		t.Error("Has() should be false after Remove")
	}
	if r.Len() != 0 {
		t.Errorf("Len() = %d, want 0", r.Len())
	}
}

func TestRegistry_Len(t *testing.T) {
	r := nav.NewRegistry()
	if r.Len() != 0 {
		t.Errorf("new registry Len() = %d, want 0", r.Len())
	}

	r.Register(testHome, mockScreen{id: "home"})
	r.Register(testQuiz, mockScreen{id: "quiz"})
	if r.Len() != 2 {
		t.Errorf("Len() = %d, want 2", r.Len())
	}
}

func TestRegistry_RegisterFactory(t *testing.T) {
	r := nav.NewRegistry()
	callCount := 0

	r.RegisterFactory(testHome, func() ui.Screen {
		callCount++
		return mockScreen{id: "home"}
	})

	screen, ok := r.Get(testHome)
	if !ok {
		t.Fatal("first Get() returned ok=false")
	}
	if screen.View() != "home" {
		t.Errorf("View() = %q, want %q", screen.View(), "home")
	}
	if callCount != 1 {
		t.Errorf("factory callCount = %d, want 1", callCount)
	}

	screen2, ok := r.Get(testHome)
	if !ok {
		t.Fatal("second Get() returned ok=false")
	}
	if screen2 != screen {
		t.Error("second Get() returned different instance")
	}
	if callCount != 1 {
		t.Errorf("factory callCount after second Get = %d, want 1", callCount)
	}
}

func TestRegistry_RegisterFactoryNotCalledUntilGet(t *testing.T) {
	r := nav.NewRegistry()
	called := false

	r.RegisterFactory(testHome, func() ui.Screen {
		called = true
		return mockScreen{id: "home"}
	})

	if called {
		t.Error("factory should not be called at registration time")
	}
}

func TestRegistry_RegisterOverwritesExisting(t *testing.T) {
	r := nav.NewRegistry()
	r.Register(testHome, mockScreen{id: "first"})
	r.Register(testHome, mockScreen{id: "second"})

	screen, ok := r.Get(testHome)
	if !ok {
		t.Fatal("Get() returned ok=false")
	}
	if screen.View() != "second" {
		t.Errorf("View() = %q, want %q", screen.View(), "second")
	}
}

func TestRegistry_RegisterFactoryOverwritesWithInstance(t *testing.T) {
	r := nav.NewRegistry()
	r.RegisterFactory(testHome, func() ui.Screen {
		return mockScreen{id: "factory"}
	})
	r.Register(testHome, mockScreen{id: "instance"})

	screen, ok := r.Get(testHome)
	if !ok {
		t.Fatal("Get() returned ok=false")
	}
	if screen.View() != "instance" {
		t.Errorf("View() = %q, want %q", screen.View(), "instance")
	}
}
