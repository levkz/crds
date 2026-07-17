package main

import (
	"github.com/alecthomas/kong"
	kongcompletion "github.com/jotaen/kong-completion"

	"crds/internal/app"
	"crds/internal/cli"
)

func main() {
	var c cli.CLI

	parser, err := kong.New(
		&c,
		kong.Name("crds"),
		kong.Description("Terminal flashcard application."),
	)
	if err != nil {
		panic(err)
	}

	kongcompletion.Register(parser)

	ctx, err := parser.Parse(nil)
	if err != nil {
		parser.FatalIfErrorf(err)
	}

	err = ctx.Run(&app.App{})
	ctx.FatalIfErrorf(err)
}
