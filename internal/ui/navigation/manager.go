package navigation

import ui "crds/internal/ui"

type Manager struct {
	stack        *Stack
	forwardStack *Stack
	modalStack   *Stack
	Current      ui.ScreenIndex
	registry     *Registry
	overlay      *ui.ScreenIndex
}

func New(initial ui.ScreenIndex) *Manager {
	return &Manager{
		stack:        NewStack(0),
		forwardStack: NewStack(0),
		modalStack:   NewStack(0),
		Current:      initial,
	}
}

func (m *Manager) Push(screen ui.ScreenIndex) PushEvent {
	event := PushEvent{From: m.Current, To: screen}
	m.stack.Push(m.Current)
	m.forwardStack = NewStack(0)
	m.Current = screen
	return event
}

func (m *Manager) Pop() (PopEvent, bool) {
	prev, ok := m.stack.Pop()
	if !ok {
		return PopEvent{}, false
	}
	event := PopEvent{From: m.Current, To: prev}
	m.forwardStack.Push(m.Current)
	m.Current = prev
	return event, true
}

func (m *Manager) Forward() (ForwardEvent, bool) {
	next, ok := m.forwardStack.Pop()
	if !ok {
		return ForwardEvent{}, false
	}
	event := ForwardEvent{From: m.Current, To: next}
	m.stack.Push(m.Current)
	m.Current = next
	return event, true
}

func (m *Manager) PushModal(screen ui.ScreenIndex) ModalPushEvent {
	event := ModalPushEvent{From: m.Current, To: screen}
	m.modalStack.Push(m.Current)
	m.Current = screen
	return event
}

func (m *Manager) DismissModal() (ModalPopEvent, bool) {
	prev, ok := m.modalStack.Pop()
	if !ok {
		return ModalPopEvent{}, false
	}
	event := ModalPopEvent{From: m.Current, To: prev}
	m.Current = prev
	return event, true
}

func (m *Manager) IsModalActive() bool {
	return !m.modalStack.IsEmpty()
}

func (m *Manager) ModalDepth() int {
	return m.modalStack.Len()
}

func (m *Manager) ShowOverlay(screen ui.ScreenIndex) OverlayShownEvent {
	event := OverlayShownEvent{Overlay: screen, Under: m.Current}
	m.overlay = &screen
	return event
}

func (m *Manager) HideOverlay() (OverlayHiddenEvent, bool) {
	if m.overlay == nil {
		return OverlayHiddenEvent{}, false
	}
	event := OverlayHiddenEvent{Overlay: *m.overlay, Under: m.Current}
	m.overlay = nil
	return event, true
}

func (m *Manager) IsOverlayActive() bool {
	return m.overlay != nil
}

func (m *Manager) Replace(screen ui.ScreenIndex) ReplaceEvent {
	event := ReplaceEvent{From: m.Current, To: screen}
	m.Current = screen
	return event
}

func (m *Manager) Reset(screen ui.ScreenIndex) ResetEvent {
	event := ResetEvent{To: screen}
	m.stack = NewStack(0)
	m.forwardStack = NewStack(0)
	m.modalStack = NewStack(0)
	m.overlay = nil
	m.Current = screen
	return event
}

func (m *Manager) SetRegistry(r *Registry) {
	m.registry = r
}

func (m *Manager) CurrentScreen() (ui.Screen, bool) {
	if m.registry == nil {
		return nil, false
	}
	return m.registry.Get(m.Current)
}

func (m *Manager) SetCurrentScreen(s ui.Screen) {
	if m.registry != nil {
		m.registry.Register(m.Current, s)
	}
}

func (m *Manager) CanGoBack() bool {
	return !m.stack.IsEmpty()
}

func (m *Manager) CanGoForward() bool {
	return !m.forwardStack.IsEmpty()
}

func (m *Manager) History() []ui.ScreenIndex {
	return m.stack.All()
}

func (m *Manager) FullHistory() []ui.ScreenIndex {
	history := m.stack.All()
	return append(history, m.Current)
}

func (m *Manager) SetMaxHistory(n int) {
	m.stack.SetLimit(n)
}

func (m *Manager) HistoryDepth() int {
	return m.stack.Len()
}

func (m *Manager) StackSize() int {
	return m.stack.Len()
}
