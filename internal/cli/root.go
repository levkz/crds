package cli

import kongcompletion "github.com/jotaen/kong-completion"

type CLI struct {
	Quiz       QuizCmd                   `cmd:"" help:"Start a quiz."`
	Sync       SyncCmd                   `cmd:"" help:"Synchronize decks and generate missing IDs."`
	Stats      StatsCmd                  `cmd:"" help:"Show learning statistics."`
	Search     SearchCmd                 `cmd:"" help:"Search vocabulary."`
	Completion kongcompletion.Completion `cmd:"" help:"Install shell completion."`
}
