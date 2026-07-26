package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"crds/internal/app"
)

type ImportCmd struct {
	Src     string `arg:"" required:"" help:"Path to a YAML file or directory."`
	Replace bool   `help:"Replace deck if it already exists."`
}

func (c *ImportCmd) Run(a *app.App) error {
	info, err := os.Stat(c.Src)
	if err != nil {
		return fmt.Errorf("import: stat %s: %w", c.Src, err)
	}

	if info.IsDir() {
		entries, err := os.ReadDir(c.Src)
		if err != nil {
			return fmt.Errorf("import: read dir %s: %w", c.Src, err)
		}
		var imported int
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			path := filepath.Join(c.Src, e.Name())
			if err := c.importFile(path, a); err != nil {
				fmt.Fprintf(os.Stderr, "warning: %v\n", err)
				continue
			}
			imported++
		}
		fmt.Printf("Imported %d file(s) from %q.\n", imported, c.Src)
		return nil
	}

	return c.importFile(c.Src, a)
}

func (c *ImportCmd) importFile(path string, a *app.App) error {
	if c.Replace {
		// Try to delete deck first — ignore "not found" errors
		deckID := strings.TrimSuffix(filepath.Base(path), ".yaml")
		if err := a.Store.DeleteDeck(deckID, a.DataDir); err != nil {
			// Only skip "not found" errors
			if !strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("import: replace %s: %w", path, err)
			}
		}
	}
	if err := a.Store.ImportDeck(path, a.DataDir); err != nil {
		return fmt.Errorf("import %s: %w", path, err)
	}
	fmt.Printf("Imported %q.\n", path)
	return nil
}
