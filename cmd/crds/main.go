package main

import (
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	kongcompletion "github.com/jotaen/kong-completion"
	"github.com/posener/complete"
	"github.com/pressly/goose/v3"

	"crds/internal/app"
	"crds/internal/cli"
	"crds/internal/storage"
)

type deckPredictor struct {
	store *storage.Store
}

func newDeckPredictor(store *storage.Store) *deckPredictor {
	return &deckPredictor{store: store}
}

func (p *deckPredictor) Predict(_ complete.Args) []string {
	names, err := p.store.ListDecks()
	if err != nil {
		return nil
	}
	return names
}

type reservePredictor struct{}

func (p *reservePredictor) Predict(_ complete.Args) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	paths, err := storage.ListReserves(filepath.Join(home, ".local", "share", "crds"))
	if err != nil {
		return nil
	}
	return paths
}

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	sharedDir := filepath.Join(home, ".local", "share", "crds")
	dataDir := filepath.Join(sharedDir, "decks")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		panic(err)
	}

	goose.SetLogger(goose.NopLogger())

	sqliteStore, err := storage.NewStore(filepath.Join(sharedDir, "crds.db"))
	if err != nil {
		panic(err)
	}
	defer sqliteStore.Close()

	stateStore := storage.NewStateStore(sharedDir)

	a := &app.App{
		Store:     sqliteStore,
		State:     stateStore,
		SharedDir: sharedDir,
		DataDir:   dataDir,
	}

	var c cli.CLI

	parser, err := kong.New(
		&c,
		kong.Name("crds"),
		kong.Description("Terminal flashcard application."),
		kong.Bind(a),
	)
	if err != nil {
		panic(err)
	}

	kongcompletion.Register(parser,
		kongcompletion.WithPredictor("deck", newDeckPredictor(sqliteStore)),
		kongcompletion.WithPredictor("reserve", &reservePredictor{}),
	)

	ctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		parser.FatalIfErrorf(err)
	}

	err = ctx.Run()
	ctx.FatalIfErrorf(err)
}
