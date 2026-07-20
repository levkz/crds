package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
)

// MsgKind classifies command result messages for typed dispatch.
type MsgKind int

const (
	MsgKindDeckList MsgKind = iota
	MsgKindDeck
	MsgKindAnswer
	MsgKindStats
)

func (k MsgKind) String() string {
	switch k {
	case MsgKindDeckList:
		return "deck-list"
	case MsgKindDeck:
		return "deck"
	case MsgKindAnswer:
		return "answer"
	case MsgKindStats:
		return "stats"
	default:
		return "unknown"
	}
}

// Dispatcher centralizes creation and execution of side-effect commands.
// Injected dependencies are accessible to all command constructors.
type Dispatcher struct {
	Decks    DeckProvider
	Progress ProgressRecorder
	Stats    StatsProvider
}

// Cmd wraps a side-effect function as a tea.Cmd for Bubble Tea dispatch.
func Cmd(f func() tea.Msg) tea.Cmd {
	return f
}

// Dispatch runs a Dispatcher-bound command, returning its tea.Cmd (a no-op when the Dispatcher is nil).
func Dispatch(d *Dispatcher, cmd func(*Dispatcher) tea.Msg) tea.Cmd {
	if d == nil {
		return nil
	}
	return func() tea.Msg {
		return cmd(d)
	}
}

// ListDecksCmd returns a command that lists available decks.
func ListDecksCmd(d *Dispatcher) tea.Cmd {
	return Dispatch(d, func(d *Dispatcher) tea.Msg {
		names, err := d.Decks.ListDecks()
		if err != nil {
			return DataErrorMsg{Kind: MsgKindDeckList, Err: err}
		}
		return DataLoadedMsg{Kind: MsgKindDeckList, Data: names}
	})
}

// LoadDeckCmd returns a command that loads a deck by name.
func LoadDeckCmd(d *Dispatcher, name string) tea.Cmd {
	return Dispatch(d, func(d *Dispatcher) tea.Msg {
		deck, err := d.Decks.LoadDeck(name)
		if err != nil {
			return DataErrorMsg{Kind: MsgKindDeck, Err: err}
		}
		return DataLoadedMsg{Kind: MsgKindDeck, Data: deck}
	})
}

// RecordAnswerCmd returns a command that persists a quiz answer.
func RecordAnswerCmd(d *Dispatcher, cardID string, grade int) tea.Cmd {
	return Dispatch(d, func(d *Dispatcher) tea.Msg {
		if err := d.Progress.RecordAnswer(cardID, grade); err != nil {
			return DataErrorMsg{Kind: MsgKindAnswer, Err: err}
		}
		return SavedMsg{Kind: MsgKindAnswer}
	})
}

// FetchStatsCmd returns a command that fetches learning statistics.
func FetchStatsCmd(d *Dispatcher) tea.Cmd {
	return Dispatch(d, func(d *Dispatcher) tea.Msg {
		stats := d.Stats.Stats()
		return StatsLoadedMsg{Stats: stats}
	})
}

// Domain command result messages

type LoadDataMsg struct {
	Kind MsgKind
}

type DataLoadedMsg struct {
	Kind MsgKind
	Data interface{}
}

type DataErrorMsg struct {
	Kind MsgKind
	Err  error
}

type SaveDataMsg struct {
	Kind MsgKind
	Data interface{}
}

type SavedMsg struct {
	Kind MsgKind
}

type SaveErrorMsg struct {
	Kind MsgKind
	Err  error
}

type StatsLoadedMsg struct {
	Stats ui.Stats
}
