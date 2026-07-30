package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	nav "crds/internal/ui/navigation"
)

// OverlayType defines which overlay is currently shown
type OverlayType int

const (
	NoOverlay OverlayType = iota
	HelpOverlay
	ConfirmOverlay
)

// Notification represents a transient message shown to the user
type Notification struct {
	Text string
}

// GlobalState holds application-wide state owned by the root model
type GlobalState struct {
	Overlay      OverlayType
	Notification *Notification
	Loading      bool
}

// Model is the root UI application model containing global state and navigation
type Model struct {
	Global GlobalState
	Config Config

	Navigator  *nav.Manager
	Dispatcher *Dispatcher

	Width  int
	Height int

	CurrentDeck    *ui.DeckData
	AllDecks       []string
	SelectedDecks []string
	AllTags        []string
	SelectedTags   []string
	AllDeckTags    map[string][]string

	AnswersRecorded bool
	PendingTarget   *ui.ScreenIndex
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		TickCmd(),
		ListDecksCmd(m.Dispatcher),
	)
}

func (m Model) WithOverlay(t OverlayType) Model {
	m.Global.Overlay = t
	return m
}

func (m Model) WithoutOverlay() Model {
	m.Global.Overlay = NoOverlay
	return m
}

func (m Model) WithNotification(text string) Model {
	m.Global.Notification = &Notification{Text: text}
	return m
}

func (m Model) WithoutNotification() Model {
	m.Global.Notification = nil
	return m
}

func (m Model) WithLoading(v bool) Model {
	m.Global.Loading = v
	return m
}

// Global state event messages

type ShowOverlayMsg struct {
	Type OverlayType
}

type HideOverlayMsg struct{}

type SetLoadingMsg struct {
	Loading bool
}

// ConfirmYesMsg is emitted when the user confirms an action dialog.
type ConfirmYesMsg struct{}

// ConfirmNoMsg is emitted when the user cancels an action dialog.
type ConfirmNoMsg struct{}
