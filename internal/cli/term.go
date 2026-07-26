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

type TermCmd struct {
	Add  TermAddCmd  `cmd:"" help:"Add a new term."`
	Rm   TermRmCmd   `cmd:"" help:"Remove a term."`
	Edit TermEditCmd `cmd:"" help:"Edit a term."`
}

type TermAddCmd struct {
	Deck string `arg:"" required:"" help:"Deck to add the term to." completion-predictor:"deck"`
}

func (c *TermAddCmd) Run(a *app.App) error {
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

type TermRmCmd struct {
	Deck   string `arg:"" required:"" help:"Deck containing the term." completion-predictor:"deck"`
	TermID string `arg:"" required:"" help:"Term ID to remove." completion-predictor:"term"`
}

func (c *TermRmCmd) Run(a *app.App) error {
	return a.Store.RemoveEntry(c.Deck, c.TermID, a.DataDir)
}

type TermEditCmd struct {
	Deck   string `arg:"" required:"" help:"Deck containing the term." completion-predictor:"deck"`
	TermID string `arg:"" required:"" help:"Term ID to edit." completion-predictor:"term"`
}

func (c *TermEditCmd) Run(a *app.App) error {
	path := filepath.Join(a.DataDir, c.Deck+".yaml")
	deck, err := parser.ParseFile(path)
	if err != nil {
		return fmt.Errorf("parse deck %q: %w", c.Deck, err)
	}

	var entry *model.Entry
	for i := range deck.Entries {
		if deck.Entries[i].ID == c.TermID {
			entry = &deck.Entries[i]
			break
		}
	}
	if entry == nil {
		return fmt.Errorf("entry %q not found in deck %q", c.TermID, c.Deck)
	}

	modified, err := editor.EditEntry(entry)
	if err != nil {
		return err
	}

	return a.Store.UpdateEntry(c.Deck, c.TermID, *modified, a.DataDir)
}
