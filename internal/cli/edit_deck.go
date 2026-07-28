package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"crds/internal/app"
	"crds/internal/editor"
	"crds/internal/model"
	"crds/internal/parser"
)

type EditDeckCmd struct {
	Deck string `arg:"" required:"" help:"Deck to edit." completion-predictor:"deck"`
}

func (c *EditDeckCmd) Run(a *app.App) error {
	path := filepath.Join(a.DataDir, c.Deck+".yaml")

	origRaw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read deck %q: %w", c.Deck, err)
	}

	origDeck, err := parser.Parse(origRaw)
	if err != nil {
		return fmt.Errorf("parse deck %q: %w", c.Deck, err)
	}

	origByID := make(map[string]model.Entry, len(origDeck.Entries))
	origByTerm := make(map[string]model.Entry, len(origDeck.Entries))
	for _, e := range origDeck.Entries {
		origByID[e.ID] = e
		origByTerm[e.Term] = e
	}

	var editedRaw []byte
	for {
		input := origRaw
		if editedRaw != nil {
			input = editedRaw
		}
		editedRaw, err = c.edit(input)
		if err != nil {
			return err
		}

		newDeck, err := parser.Parse(editedRaw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			choice := c.prompt("Discard changes, Continue editing, Save anyway", "d", "c", "s")
			switch choice {
			case "d":
				fmt.Println("Changes discarded.")
				return nil
			case "c":
				continue
			case "s":
				if err := os.WriteFile(path, editedRaw, 0644); err != nil {
					return fmt.Errorf("write deck: %w", err)
				}
				fmt.Println("Saved (parse errors may cause sync warnings).")
				return nil
			}
		}

		newByID := make(map[string]model.Entry, len(newDeck.Entries))
		newByTerm := make(map[string]model.Entry, len(newDeck.Entries))
		for _, e := range newDeck.Entries {
			newByID[e.ID] = e
			newByTerm[e.Term] = e
		}

		var migrations []struct{ oldID, newID string }
		for term, newEntry := range newByTerm {
			origEntry, ok := origByTerm[term]
			if !ok || origEntry.ID == newEntry.ID {
				continue
			}
			fmt.Fprintf(os.Stderr, "Entry %q changed ID from %q to %q (term unchanged).\n", term, origEntry.ID, newEntry.ID)
			choice := c.prompt("Migrate statistics to new ID, Create as new entry", "m", "c")
			switch choice {
			case "m":
				migrations = append(migrations, struct{ oldID, newID string }{origEntry.ID, newEntry.ID})
			case "c":
			}
		}

		var deleted []model.Entry
		for _, origEntry := range origByID {
			if _, exists := newByID[origEntry.ID]; exists {
				continue
			}
			if _, exists := newByTerm[origEntry.Term]; exists {
				continue
			}
			deleted = append(deleted, origEntry)
		}

		if len(deleted) > 0 {
			fmt.Fprintf(os.Stderr, "%d entr(ies) were deleted.\n", len(deleted))
			choice := c.prompt("Clear cache for all, Revert all deletions, Review each", "c", "r", "e")
			switch choice {
			case "c":
				for _, e := range deleted {
					if err := a.Store.RemoveEntry(c.Deck, e.ID, a.DataDir); err != nil {
						return fmt.Errorf("remove entry %q: %w", e.ID, err)
					}
				}
			case "r":
				for _, e := range deleted {
					if err := a.Store.AddEntry(c.Deck, e, a.DataDir); err != nil {
						return fmt.Errorf("re-add entry %q: %w", e.ID, err)
					}
				}
			case "e":
				for _, e := range deleted {
					fmt.Fprintf(os.Stderr, "Entry %q (term: %q) was deleted.\n", e.ID, e.Term)
					choice := c.prompt("Clear cache, Revert deletion, Skip", "c", "r", "s")
					switch choice {
					case "c":
						if err := a.Store.RemoveEntry(c.Deck, e.ID, a.DataDir); err != nil {
							return fmt.Errorf("remove entry %q: %w", e.ID, err)
						}
					case "r":
						if err := a.Store.AddEntry(c.Deck, e, a.DataDir); err != nil {
							return fmt.Errorf("re-add entry %q: %w", e.ID, err)
						}
					case "s":
					}
				}
			}
		}

		if err := os.WriteFile(path, editedRaw, 0644); err != nil {
			return fmt.Errorf("write deck: %w", err)
		}

		for _, m := range migrations {
			if err := a.Store.ReplaceEntryID(c.Deck, m.oldID, m.newID, a.DataDir); err != nil {
				return fmt.Errorf("migrate entry %q → %q: %w", m.oldID, m.newID, err)
			}
		}

		fmt.Printf("Deck %q updated.\n", c.Deck)
		return nil
	}
}

func (c *EditDeckCmd) edit(raw []byte) ([]byte, error) {
	result, err := editor.Edit(string(raw))
	if err != nil {
		return nil, err
	}
	return []byte(result), nil
}

func (c *EditDeckCmd) prompt(message string, options ...string) string {
	valid := make(map[string]bool, len(options))
	for _, o := range options {
		valid[o] = true
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprintf(os.Stderr, "%s [%s]: ", message, strings.Join(options, "/"))
		if !scanner.Scan() {
			return options[0]
		}
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if valid[answer] {
			return answer
		}
		fmt.Fprintf(os.Stderr, "Invalid choice. Valid options: %s\n", strings.Join(options, ", "))
	}
}
