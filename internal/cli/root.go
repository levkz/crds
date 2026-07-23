package cli

import (
	"log"
	"os"
	"path/filepath"

	kongcompletion "github.com/jotaen/kong-completion"
	"github.com/pressly/goose/v3"

	"crds/internal/app"
	uiapp "crds/internal/ui/app"
	"crds/internal/storage"
)

type CLI struct {
	Debug      bool                      `help:"Enable debug output."`
	Quiz       QuizCmd                   `cmd:"" help:"Start a quiz."`
	Sync       SyncCmd                   `cmd:"" help:"Synchronize decks and generate missing IDs."`
	Stats      StatsCmd                  `cmd:"" help:"Show learning statistics."`
	Search     SearchCmd                 `cmd:"" help:"Search vocabulary."`
	Completion kongcompletion.Completion `cmd:"" help:"Install shell completion."`
}

func (c *CLI) Run(a *app.App) error {
	if !c.Debug {
		goose.SetLogger(goose.NopLogger())
	}

	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	sharedDir := filepath.Join(home, ".local", "share", "crds")
	dataDir := filepath.Join(sharedDir, "decks")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	stateStore := storage.NewStateStore(sharedDir)

	// Open the SQLite database
	dbPath := filepath.Join(sharedDir, "crds.db")
	sqliteStore, err := storage.NewStore(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer sqliteStore.Close()

	// Sync decks from YAML to SQLite cache (only if files changed)
	if err := sqliteStore.SyncDecks(dataDir); err != nil {
		log.Fatalf("failed to sync decks: %v", err)
	}

	deps := uiapp.Dependencies{
		Decks:    sqliteStore,
		Progress: sqliteStore,
		Stats:    sqliteStore,
		State:    stateStore,
	}

	return uiapp.RunWithDefaults(deps)
}
