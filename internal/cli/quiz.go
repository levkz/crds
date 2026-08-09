package cli

import (
	"fmt"
	"os"

	"crds/internal/app"
	uiapp "crds/internal/ui/app"
)

type QuizCmd struct {
	Deck string `arg:"" optional:"" help:"Deck to quiz." completion-predictor:"deck"`

	Reverse bool `help:"Reverse the quiz direction."`

	Limit int `short:"n" default:"20" help:"Maximum number of cards."`
}

func (q *QuizCmd) Run(a *app.App) error {
	if err := a.Store.SyncDecks(a.DataDir); err != nil {
		return err
	}

	if q.Deck != "" {
		if err := q.preSelectDeck(a); err != nil {
			return err
		}
	}

	if q.Limit != 20 {
		fmt.Fprintf(os.Stderr, "warning: --limit is not yet implemented in the TUI\n")
	}
	if q.Reverse {
		fmt.Fprintf(os.Stderr, "warning: --reverse is not yet implemented in the TUI\n")
	}

	deps := uiapp.Dependencies{
		Decks:    a.Store,
		Progress: a.Store,
		Stats:    a.Store,
		State:    a.State,
		Due:      a.Store,
	}

	return uiapp.RunWithDefaults(deps)
}

func (q *QuizCmd) preSelectDeck(a *app.App) error {
	decks, err := a.Store.ListDecks()
	if err != nil {
		return fmt.Errorf("list decks: %w", err)
	}

	found := false
	for _, d := range decks {
		if d == q.Deck {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("deck %q not found", q.Deck)
	}

	state, err := a.State.Load()
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}
	state.SelectedDecks = []string{q.Deck}
	return a.State.Save(state)
}
