package cli

type DeckCmd struct {
	List   ListCmd     `cmd:"" help:"List all decks with entry counts."`
	Import ImportCmd   `cmd:"" help:"Import a deck from a YAML file."`
	Export ExportCmd   `cmd:"" help:"Export a deck to a YAML file."`
	Delete DeleteCmd   `cmd:"" help:"Delete a deck."`
	Search SearchCmd   `cmd:"" help:"Search vocabulary in a deck."`
	Edit   EditDeckCmd `cmd:"" help:"Edit a deck by opening its YAML file."`
	Term   TermCmd     `cmd:"" help:"Manage individual terms in a deck."`
	Tag    TagCmd      `cmd:"" help:"Manage tags on terms."`
}
