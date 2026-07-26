package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/storage"
	"crds/internal/ui"
	"crds/internal/ui/theme"
)

// MsgKind classifies command result messages for typed dispatch.
type MsgKind int

const (
	MsgKindDeckList MsgKind = iota
	MsgKindDeck
	MsgKindAnswer
	MsgKindStats
	MsgKindState
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
	case MsgKindState:
		return "state"
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
	State    *storage.StateStore
	Sessions SessionManager
	Typing   TypingRecorder
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

// LoadSelectedDecksCmd loads multiple decks and merges them into a single DeckData.
func LoadSelectedDecksCmd(d *Dispatcher, names []string) tea.Cmd {
	return Dispatch(d, func(d *Dispatcher) tea.Msg {
		var decks []ui.DeckData
		for _, name := range names {
			deck, err := d.Decks.LoadDeck(name)
			if err != nil {
				return DataErrorMsg{Kind: MsgKindDeck, Err: fmt.Errorf("load %q: %w", name, err)}
			}
			decks = append(decks, deck)
		}
		merged := mergeDecks(decks)
		return DataLoadedMsg{Kind: MsgKindDeck, Data: merged}
	})
}

// mergeDecks combines multiple DeckData into one by concatenating cards.
func mergeDecks(decks []ui.DeckData) ui.DeckData {
	if len(decks) == 0 {
		return ui.DeckData{}
	}
	if len(decks) == 1 {
		return decks[0]
	}
	var nameParts []string
	var allCards []ui.CardData
	for _, d := range decks {
		nameParts = append(nameParts, d.Name)
		allCards = append(allCards, d.Cards...)
	}
	return ui.DeckData{
		Name:  strings.Join(nameParts, " + "),
		Cards: allCards,
	}
}

// SaveStateCmd persists the selected decks and active theme.
func SaveStateCmd(d *Dispatcher, selected []string) tea.Msg {
	if d.State == nil {
		return nil
	}
	err := d.State.Save(&storage.State{
		SelectedDecks: selected,
		Theme:         theme.CurrentName(),
	})
	if err != nil {
		return DataErrorMsg{Kind: MsgKindState, Err: fmt.Errorf("saving state: %w", err)}
	}
	return SavedMsg{Kind: MsgKindState}
}

// RecordAnswerCmd returns a command that persists a quiz answer.
// If the msg has typing details, they are recorded via TypingRecorder.
func RecordAnswerCmd(d *Dispatcher, msg ui.SaveAnswerMsg) tea.Cmd {
	return Dispatch(d, func(d *Dispatcher) tea.Msg {
		if d.Typing != nil && (msg.UserInput != "" || msg.CorrectAnswer != "") {
			_, err := d.Typing.RecordAnswerFull(0, msg.DeckID, msg.CardID, msg.Grade, msg.Reverse, msg.UserInput, msg.CorrectAnswer, msg.Similarity)
			if err != nil {
				return DataErrorMsg{Kind: MsgKindAnswer, Err: err}
			}
		} else {
			if err := d.Progress.RecordAnswer(msg.DeckID, msg.CardID, msg.Grade, msg.Reverse); err != nil {
				return DataErrorMsg{Kind: MsgKindAnswer, Err: err}
			}
		}
		return SavedMsg{Kind: MsgKindAnswer}
	})
}

// ResetSessionCmd returns a command that resets the current session.
func ResetSessionCmd(d *Dispatcher) tea.Cmd {
	return Dispatch(d, func(d *Dispatcher) tea.Msg {
		if d.Sessions != nil {
			_ = d.Sessions.ResetSession()
		}
		return nil
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
