package cli

type SearchCmd struct {
	Query string `arg:"" required:"" help:"Search query."`
}
