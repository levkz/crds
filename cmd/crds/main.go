package main

import (
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	kongcompletion "github.com/jotaen/kong-completion"
	"github.com/posener/complete"

	"crds/internal/cli"
	"crds/internal/storage"
)

type deckPredictor struct {
	store *storage.DeckStore
}

func newDeckPredictor(store *storage.DeckStore) *deckPredictor {
	return &deckPredictor{store: store}
}

func (p *deckPredictor) Predict(_ complete.Args) []string {
	names, err := p.store.ListDecks()
	if err != nil {
		return nil
	}
	return names
}

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dataDir := filepath.Join(home, ".local", "share", "crds", "decks")
	deckStore := storage.NewDeckStore(dataDir)

	var c cli.CLI

	parser, err := kong.New(
		&c,
		kong.Name("crds"),
		kong.Description("Terminal flashcard application."),
	)
	if err != nil {
		panic(err)
	}

	kongcompletion.Register(parser, kongcompletion.WithPredictor("deck", newDeckPredictor(deckStore)))

	ctx, err := parser.Parse(nil)
	if err != nil {
		parser.FatalIfErrorf(err)
	}

	err = ctx.Run()
	ctx.FatalIfErrorf(err)
}
