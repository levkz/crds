package app

import (
	"fmt"
	"strings"
	"time"

	"crds/internal/stats"
	"crds/internal/storage"
	"crds/internal/ui"
	"crds/internal/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
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
	Due      DueProvider
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
// Mappings union with later decks overriding; the language defaults come from
// the first deck's language.
func mergeDecks(decks []ui.DeckData) ui.DeckData {
	if len(decks) == 0 {
		return ui.DeckData{}
	}
	if len(decks) == 1 {
		return decks[0]
	}
	var nameParts []string
	var allCards []ui.CardData
	mappings := map[string]string{}
	for i, d := range decks {
		nameParts = append(nameParts, d.Name)
		allCards = append(allCards, d.Cards...)
		if i == 0 {
			mappings = d.InputMappings
		} else if len(d.InputMappings) > 0 {
			merged := map[string]string{}
			for k, v := range mappings {
				merged[k] = v
			}
			for k, v := range d.InputMappings {
				merged[k] = v
			}
			mappings = merged
		}
	}
	return ui.DeckData{
		Name:          strings.Join(nameParts, " + "),
		Language:      decks[0].Language,
		InputMappings: mappings,
		Cards:         allCards,
	}
}

// LoadDueProgressMsg carries the refreshed review queue (due entry IDs) and
// per-card aggregate progress for the selection. It is returned by
// LoadDueProgressCmd, which is invoked after every recorded answer so the
// statistics due count and the due-mode queue stay fresh without reshuffling
// the active quiz session.

type LoadDueProgressMsg struct {
	Due      []string
	Progress map[string]stats.EntryProgress
}

// LoadDueProgressCmd refreshes the due queue and per-entry progress for a
// deck/tag selection.
func LoadDueProgressCmd(d *Dispatcher, deckIDs, tags []string) tea.Cmd {
	return Dispatch(d, func(d *Dispatcher) tea.Msg {
		msg := LoadDueProgressMsg{Progress: make(map[string]stats.EntryProgress)}
		for _, name := range deckIDs {
			prog, err := d.Stats.EntryProgress(name)
			if err != nil {
				continue
			}
			for k, v := range prog {
				msg.Progress[k] = v
			}
		}
		if d.Due != nil {
			due, err := d.Due.DueForSelection(deckIDs, tags, time.Now())
			if err != nil {
				return DataErrorMsg{Kind: MsgKindStats, Err: err}
			}
			msg.Due = due
		}
		return msg
	})
}

// SaveStateCmd persists the selected decks, tags, active theme, and quiz mode.
func SaveStateCmd(d *Dispatcher, selected []string, mode ui.QuizMode, selectedTags ...string) tea.Msg {
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
		QuizMode:      mode.String(),
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

// FetchStatsCmd returns a command that fetches learning statistics for the
// current deck/tag selection (an empty selection covers all decks).
func FetchStatsCmd(d *Dispatcher, deckIDs, tags []string) tea.Cmd {
	return Dispatch(d, func(d *Dispatcher) tea.Msg {
		g, err := d.Stats.Summary()
		if err != nil {
			return DataErrorMsg{Kind: MsgKindStats, Err: err}
		}
		sel, err := d.Stats.SelectionSummary(deckIDs, tags)
		if err != nil {
			return DataErrorMsg{Kind: MsgKindStats, Err: err}
		}
		hist, err := d.Stats.SelectionHistory(deckIDs, tags)
		if err != nil {
			return DataErrorMsg{Kind: MsgKindStats, Err: err}
		}
		return StatsLoadedMsg{Stats: g, SelectionStats: &sel, SelectionHistory: hist}
	})
}

// FetchWordStatsCmd returns a command that fetches per-word statistics.
func FetchWordStatsCmd(d *Dispatcher, entryID string) tea.Cmd {
	return Dispatch(d, func(d *Dispatcher) tea.Msg {
		ws, err := d.Stats.WordStats(entryID)
		if err != nil {
			return DataErrorMsg{Kind: MsgKindStats, Err: err}
		}
		hist, err := d.Stats.WordHistory(entryID)
		if err != nil {
			return DataErrorMsg{Kind: MsgKindStats, Err: err}
		}
		return ui.WordStatsLoadedMsg{EntryID: entryID, Stats: ws, History: hist}
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
	Stats            stats.Summary
	SelectionStats   *stats.Summary
	SelectionHistory []stats.DayPoint
}
