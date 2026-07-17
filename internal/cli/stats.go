package cli

type StatsCmd struct {
	Deck string `arg:"" optional:"" help:"Show stats for a single deck." completion-predictor:"deck"`
}
