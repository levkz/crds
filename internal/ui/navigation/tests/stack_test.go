package navigation_test

import (
	"testing"

	ui "crds/internal/ui"
	nav "crds/internal/ui/navigation"
)

func TestStack_PushPop(t *testing.T) {
	s := nav.NewStack(0)
	s.Push(testHome)
	s.Push(testQuiz)

	if s.Len() != 2 {
		t.Fatalf("stack.Len() = %d, want %d", s.Len(), 2)
	}

	screen, ok := s.Pop()
	if !ok {
		t.Fatal("Pop() returned ok=false, want true")
	}
	if screen != testQuiz {
		t.Errorf("Pop() = %d, want %d", screen, testQuiz)
	}

	screen, ok = s.Pop()
	if !ok {
		t.Fatal("Pop() returned ok=false, want true")
	}
	if screen != testHome {
		t.Errorf("Pop() = %d, want %d", screen, testHome)
	}

	if !s.IsEmpty() {
		t.Error("stack should be empty after popping all items")
	}
}

func TestStack_PopEmpty(t *testing.T) {
	s := nav.NewStack(0)
	_, ok := s.Pop()
	if ok {
		t.Fatal("Pop() on empty stack returned ok=true")
	}
}

func TestStack_Peek(t *testing.T) {
	s := nav.NewStack(0)
	s.Push(testHome)
	s.Push(testQuiz)

	screen, ok := s.Peek()
	if !ok {
		t.Fatal("Peek() returned ok=false")
	}
	if screen != testQuiz {
		t.Errorf("Peek() = %d, want %d", screen, testQuiz)
	}
	if s.Len() != 2 {
		t.Errorf("Peek() should not modify stack, Len() = %d", s.Len())
	}
}

func TestStack_PeekEmpty(t *testing.T) {
	s := nav.NewStack(0)
	_, ok := s.Peek()
	if ok {
		t.Fatal("Peek() on empty stack returned ok=true")
	}
}

func TestStack_IsEmpty(t *testing.T) {
	s := nav.NewStack(0)
	if !s.IsEmpty() {
		t.Error("new stack should be empty")
	}
	s.Push(testHome)
	if s.IsEmpty() {
		t.Error("stack with element should not be empty")
	}
}

func TestStack_DepthLimit(t *testing.T) {
	s := nav.NewStack(2)
	s.Push(testHome)
	s.Push(testQuiz)
	s.Push(testSearch)

	if s.Len() != 2 {
		t.Fatalf("stack.Len() = %d, want %d", s.Len(), 2)
	}

	all := s.All()
	want := []ui.ScreenIndex{testQuiz, testSearch}
	if len(all) != len(want) {
		t.Fatalf("stack.All() length = %d, want %d", len(all), len(want))
	}
	for i, v := range all {
		if v != want[i] {
			t.Errorf("stack.All()[%d] = %d, want %d", i, v, want[i])
		}
	}
}

func TestStack_All(t *testing.T) {
	s := nav.NewStack(0)
	s.Push(testHome)
	s.Push(testQuiz)

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("stack.All() length = %d", len(all))
	}

	s.Push(testSearch)
	if len(all) != 2 {
		t.Errorf("stack.All() should return a copy, length changed to %d", len(all))
	}
}

func TestStack_ZeroLimit(t *testing.T) {
	s := nav.NewStack(0)
	for i := 0; i < 100; i++ {
		s.Push(testHome)
	}
	if s.Len() != 100 {
		t.Errorf("zero-limit stack should grow unbounded, Len() = %d", s.Len())
	}
}
