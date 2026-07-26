package cli

import (
	"fmt"
	"path/filepath"

	"go.yaml.in/yaml/v3"

	"crds/internal/app"
	"crds/internal/editor"
	"crds/internal/model"
	"crds/internal/parser"
)

type EditCmd struct {
	Deck    string `arg:"" required:"" help:"Deck containing the entry." completion-predictor:"deck"`
	EntryID string `arg:"" optional:"" help:"Entry ID to edit (omit to create a new entry)."`
}

func (c *EditCmd) Run(a *app.App) error {
	if c.EntryID == "" {
		return c.createEntry(a)
	}
	return c.editEntry(a)
}

func (c *EditCmd) createEntry(a *app.App) error {
	content, err := editor.Edit(editor.EntryTemplate())
	if err != nil {
		return err
	}
	var entry model.Entry
	if err := yaml.Unmarshal([]byte(content), &entry); err != nil {
		return fmt.Errorf("parse entry: %w", err)
	}
	return a.Store.AddEntry(c.Deck, entry, a.DataDir)
}

func (c *EditCmd) editEntry(a *app.App) error {
	path := filepath.Join(a.DataDir, c.Deck+".yaml")
	deck, err := parser.ParseFile(path)
	if err != nil {
		return fmt.Errorf("parse deck %q: %w", c.Deck, err)
	}

	var entry *model.Entry
	for i := range deck.Entries {
		if deck.Entries[i].ID == c.EntryID {
			entry = &deck.Entries[i]
			break
		}
	}
	if entry == nil {
		return fmt.Errorf("entry %q not found in deck %q", c.EntryID, c.Deck)
	}

	modified, err := editor.EditEntry(entry)
	if err != nil {
		return err
	}

	return a.Store.UpdateEntry(c.Deck, c.EntryID, *modified, a.DataDir)
}
