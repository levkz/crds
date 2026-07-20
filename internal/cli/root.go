package cli

import (
	"os"
	"path/filepath"

	kongcompletion "github.com/jotaen/kong-completion"

	"crds/internal/app"
	uiapp "crds/internal/ui/app"
	"crds/internal/storage"
)

type CLI struct {
	Quiz       QuizCmd                   `cmd:"" help:"Start a quiz."`
	Sync       SyncCmd                   `cmd:"" help:"Synchronize decks and generate missing IDs."`
	Stats      StatsCmd                  `cmd:"" help:"Show learning statistics."`
	Search     SearchCmd                 `cmd:"" help:"Search vocabulary."`
	Completion kongcompletion.Completion `cmd:"" help:"Install shell completion."`
}

func (c *CLI) Run(a *app.App) error {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dataDir := filepath.Join(home, ".local", "share", "crds", "decks")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	deckStore := storage.NewDeckStore(dataDir)
	progressStore := storage.NewProgressStore()

	deps := uiapp.Dependencies{
		Decks:    deckStore,
		Progress: progressStore,
		Stats:    progressStore,
	}

	return uiapp.RunWithDefaults(deps)
}
