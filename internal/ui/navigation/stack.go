package navigation

import ui "crds/internal/ui"

type Stack struct {
	screens []ui.ScreenIndex
	limit   int
}

func NewStack(limit int) *Stack {
	return &Stack{limit: limit}
}

func (s *Stack) Push(screen ui.ScreenIndex) {
	if s.limit > 0 && len(s.screens) >= s.limit {
		s.screens = s.screens[1:]
	}
	s.screens = append(s.screens, screen)
}

func (s *Stack) Pop() (ui.ScreenIndex, bool) {
	if len(s.screens) == 0 {
		return 0, false
	}
	idx := len(s.screens) - 1
	screen := s.screens[idx]
	s.screens = s.screens[:idx]
	return screen, true
}

func (s *Stack) Peek() (ui.ScreenIndex, bool) {
	if len(s.screens) == 0 {
		return 0, false
	}
	return s.screens[len(s.screens)-1], true
}

func (s *Stack) Len() int {
	return len(s.screens)
}

func (s *Stack) IsEmpty() bool {
	return len(s.screens) == 0
}

func (s *Stack) All() []ui.ScreenIndex {
	out := make([]ui.ScreenIndex, len(s.screens))
	copy(out, s.screens)
	return out
}

func (s *Stack) SetLimit(n int) {
	s.limit = n
	if n > 0 && len(s.screens) > n {
		s.screens = s.screens[len(s.screens)-n:]
	}
}
