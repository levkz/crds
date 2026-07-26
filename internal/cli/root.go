package cli

import (
	"github.com/alecthomas/kong"
	kongcompletion "github.com/jotaen/kong-completion"

	"crds/internal/app"
	uiapp "crds/internal/ui/app"
)

type CLI struct {
	Debug      bool                      `help:"Enable debug output."`
	Quiz       QuizCmd                   `cmd:"" help:"Start a quiz."`
	Sync       SyncCmd                   `cmd:"" help:"Synchronize decks and generate missing IDs."`
	Stats      StatsCmd                  `cmd:"" help:"Show learning statistics."`
	Search     SearchCmd                 `cmd:"" help:"Search vocabulary."`
	Import     ImportCmd                 `cmd:"" help:"Import a deck from a YAML file."`
	Export     ExportCmd                 `cmd:"" help:"Export a deck to a YAML file."`
	Delete     DeleteCmd                 `cmd:"" help:"Delete a deck."`
	Reserve    ReserveCmd                `cmd:"" help:"Create a backup/reserve copy."`
	Revert     RevertCmd                 `cmd:"" help:"Revert from a reserve copy."`
	Edit       EditCmd                   `cmd:"" help:"Edit a deck entry."`
	Completion kongcompletion.Completion `cmd:"" help:"Install shell completion."`
}

func (c *CLI) Run(a *app.App, ctx *kong.Context) error {
	if ctx.Selected() != nil {
		return nil
	}

	if err := a.Store.SyncDecks(a.DataDir); err != nil {
		return err
	}

	deps := uiapp.Dependencies{
		Decks:    a.Store,
		Progress: a.Store,
		Stats:    a.Store,
		State:    a.State,
	}

	return uiapp.RunWithDefaults(deps)
}
