package cli

type StateCmd struct {
	Reserve ReserveCmd `cmd:"" help:"Create a backup/reserve copy."`
	Revert  RevertCmd  `cmd:"" help:"Revert from a reserve copy."`
	Sync    SyncCmd    `cmd:"" help:"Synchronize decks and generate missing IDs."`
}
