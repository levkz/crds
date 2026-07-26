package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"crds/internal/app"
	"crds/internal/storage"
)

type SearchCmd struct {
	Query string   `arg:"" optional:"" help:"Search query (empty for all entries)."`
	Deck  []string `help:"Deck(s) to search in (repeatable, defaults to all)."`
	Tags  []string `help:"Tags to filter by (repeatable, AND logic)."`
}

func (c *SearchCmd) Run(a *app.App) error {
	results, err := a.Store.Search(context.Background(), c.Query, c.Deck, c.Tags)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No matches found.")
		return nil
	}

	type deckGroup struct {
		Name    string
		Entries []storage.SearchResult
	}
	groups := make(map[string]*deckGroup)
	var deckOrder []string
	for _, r := range results {
		g, ok := groups[r.DeckID]
		if !ok {
			g = &deckGroup{Name: r.DeckName}
			groups[r.DeckID] = g
			deckOrder = append(deckOrder, r.DeckID)
		}
		g.Entries = append(g.Entries, r)
	}

	fmt.Printf("%d match(es):\n\n", len(results))
	for _, dID := range deckOrder {
		g := groups[dID]
		fmt.Printf("=== %s (%s) — %d match(es) ===\n", g.Name, dID, len(g.Entries))
		sort.Slice(g.Entries, func(i, j int) bool {
			return g.Entries[i].Term < g.Entries[j].Term
		})
		for _, r := range g.Entries {
			tags := ""
			if len(r.Tags) > 0 {
				tags = " [" + strings.Join(r.Tags, ",") + "]"
			}
			translations := strings.Join(r.Translations, ", ")
			fmt.Printf("  %s%s  → %s\n", r.Term, tags, translations)
			if r.Notes != "" {
				fmt.Printf("         notes: %s\n", r.Notes)
			}
		}
		fmt.Println()
	}

	return nil
}
