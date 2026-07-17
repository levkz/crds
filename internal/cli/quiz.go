package cli

import (
	"fmt"

	"crds/internal/app"
)

type QuizCmd struct {
	Deck string `arg:"" optional:"" help:"Deck to quiz." completion-predictor:"deck"`

	Reverse bool `help:"Reverse the quiz direction."`

	Limit int `short:"n" default:"20" help:"Maximum number of cards."`
}

func (q *QuizCmd) Run(a *app.App) error {
	fmt.Printf("Quiz deck=%q reverse=%v limit=%d\n",
		q.Deck,
		q.Reverse,
		q.Limit,
	)

	return nil
}
