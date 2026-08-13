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
	"crds/internal/config"
	"crds/internal/storage"
	"crds/internal/ui/theme"
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

type entryPredictor struct {
	store *storage.Store
}

func newEntryPredictor(store *storage.Store) *entryPredictor {
	return &entryPredictor{store: store}
}

type themePredictor struct{}

func (p *themePredictor) Predict(_ complete.Args) []string {
	files, err := config.DiscoverThemeFiles()
	if err != nil {
		return nil
	}
	names := make([]string, len(files))
	for i, tf := range files {
		names[i] = tf.Name
	}
	return names
}

type presetPredictor struct{}

func (p *presetPredictor) Predict(_ complete.Args) []string {
	return theme.BuiltinNames()
}

func (p *entryPredictor) Predict(args complete.Args) []string {
	if len(args.Completed) == 0 {
		return nil
	}
	deck := args.Completed[len(args.Completed)-1]
	entries, err := p.store.LoadDeck(deck)
	if err != nil {
		return nil
	}
	ids := make([]string, len(entries.Cards))
	for i, card := range entries.Cards {
		ids[i] = card.ID
	}
	return ids
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
	defer func() { _ = sqliteStore.Close() }()

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
		kongcompletion.WithPredictor("term", newEntryPredictor(sqliteStore)),
		kongcompletion.WithPredictor("theme", &themePredictor{}),
		kongcompletion.WithPredictor("preset", &presetPredictor{}),
	)

	ctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		parser.FatalIfErrorf(err)
	}

	err = ctx.Run()
	ctx.FatalIfErrorf(err)
}
