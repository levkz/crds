package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/stats"
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
	MsgKindTags
	MsgKindDeckTags
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
	Stats    stats.Provider
	State    *storage.StateStore
	Sessions SessionManager
	Typing   TypingRecorder
	Tags     TagProvider
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

// DeckWithProgressMsg carries deck data plus per-card progress for sorting.
type DeckWithProgressMsg struct {
	Deck     ui.DeckData
	Progress map[string]stats.EntryProgress
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

		// Load progress for smart sorting
		if d.Stats != nil && len(names) > 0 {
			allProgress := make(map[string]stats.EntryProgress)
			for _, name := range names {
				prog, err := d.Stats.EntryProgress(name)
				if err != nil {
					continue
				}
				for k, v := range prog {
					allProgress[k] = v
				}
			}
			return DataLoadedMsg{Kind: MsgKindDeck, Data: DeckWithProgressMsg{
				Deck:     merged,
				Progress: allProgress,
			}}
		}

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

// SaveStateCmd persists the selected decks, tags, and active theme.
func SaveStateCmd(d *Dispatcher, selected []string, selectedTags ...string) tea.Msg {
	if d.State == nil {
		return nil
	}
	var tags []string
	if len(selectedTags) > 0 {
		tags = selectedTags
	}
	err := d.State.Save(&storage.State{
		SelectedDecks: selected,
		SelectedTags:  tags,
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
			_, err := d.Typing.RecordAnswerFull(0, msg.DeckID, msg.CardID, int(msg.Grade), msg.Reverse, msg.UserInput, msg.CorrectAnswer, msg.Similarity)
			if err != nil {
				return DataErrorMsg{Kind: MsgKindAnswer, Err: err}
			}
		} else {
			if err := d.Progress.RecordAnswer(msg.DeckID, msg.CardID, int(msg.Grade), msg.Reverse); err != nil {
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

// ListAllTagsCmd returns a command that fetches all unique tags.
func ListAllTagsCmd(d *Dispatcher) tea.Cmd {
	return Dispatch(d, func(d *Dispatcher) tea.Msg {
		tags, err := d.Tags.ListAllTags()
		if err != nil {
			return DataErrorMsg{Kind: MsgKindTags, Err: err}
		}
		return DataLoadedMsg{Kind: MsgKindTags, Data: tags}
	})
}

// LoadAllDeckTagsCmd returns a command that loads tags for every deck (deck→tags map).
func LoadAllDeckTagsCmd(d *Dispatcher) tea.Cmd {
	return Dispatch(d, func(d *Dispatcher) tea.Msg {
		deckIDs, err := d.Decks.ListDecks()
		if err != nil {
			return DataErrorMsg{Kind: MsgKindDeckTags, Err: err}
		}
		result := make(map[string][]string, len(deckIDs))
		for _, id := range deckIDs {
			tags, err := d.Tags.ListDeckTags(id)
			if err != nil {
				continue
			}
			result[id] = tags
		}
		return DataLoadedMsg{Kind: MsgKindDeckTags, Data: result}
	})
}

// FetchStatsCmd returns a command that fetches learning statistics.
func FetchStatsCmd(d *Dispatcher) tea.Cmd {
	return Dispatch(d, func(d *Dispatcher) tea.Msg {
		s, err := d.Stats.Summary()
		if err != nil {
			return DataErrorMsg{Kind: MsgKindStats, Err: err}
		}
		return StatsLoadedMsg{Stats: s}
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
	Stats stats.Summary
}
