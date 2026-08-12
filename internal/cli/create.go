package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"

	"crds/internal/app"
	"crds/internal/model"
	"crds/internal/parser"
)

type CreateCmd struct {
	Deck string `arg:"" required:"" help:"Name (and ID) of the new deck."`
	From string `short:"F" required:"" help:"Source language for terms."`
	To   string `short:"T" required:"" help:"Target language for translations."`
	Edit bool   `help:"Open the new deck in the editor after creating it."`
}

func (c *CreateCmd) Run(a *app.App) error {
	path := filepath.Join(a.DataDir, c.Deck+".yaml")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("create: deck %q already exists at %s", c.Deck, path)
	}

	deck := &model.Deck{
		ID:                  c.Deck,
		Name:                c.Deck,
		Language:            c.From,
		TranslationLanguage: c.To,
	}
	data, err := yaml.Marshal(deck)
	if err != nil {
		return fmt.Errorf("create: render deck: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("create: write %s: %w", path, err)
	}

	if _, err := parser.ParseFile(path); err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("create: %w", err)
	}

	if err := a.Store.SyncDecks(a.DataDir); err != nil {
		return fmt.Errorf("create: sync: %w", err)
	}
	fmt.Printf("Created deck %q (%s → %s).\n", c.Deck, c.From, c.To)

	if c.Edit {
		return (&EditDeckCmd{Deck: c.Deck}).Run(a)
	}
	return nil
}