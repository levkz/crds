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
	Stats      StatsCmd                  `cmd:"" help:"Show learning statistics."`
	Deck       DeckCmd                   `cmd:"" help:"Deck operations (import, export, search, delete, edit, term)."`
	Theme      ThemeCmd                  `cmd:"" help:"Theme operations (add, delete, edit, list)."`
	State      StateCmd                  `cmd:"" help:"State management (reserve, revert, sync)."`
	Profile    ProfileCmd                `cmd:"" help:"Profile operations (export, import)."`
	Ai         AiCmd                     `cmd:"" help:"AI agent (interpret, fill, add)."`
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
		Due:      a.Store,
	}

	return uiapp.RunWithDefaults(deps)
}
