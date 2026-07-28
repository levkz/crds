package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/mattn/go-isatty"

	"crds/internal/app"
	"crds/internal/storage"
)

type SearchCmd struct {
	Query string   `arg:"" optional:"" help:"Search query (empty for all entries)."`
	Deck  []string `help:"Deck(s) to search in (repeatable, defaults to all)." completion-predictor:"deck"`
	Tags  []string `help:"Tags to filter by (repeatable, AND logic)."`
	Color string   `enum:"auto,always,never" default:"auto" help:"When to use color/highlighting."`
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

	var open, close string
	if c.Color != "never" && c.Query != "" {
		if c.Color == "always" || isatty.IsTerminal(os.Stdout.Fd()) {
			open = grepMatchStyle()
			close = "\033[0m"
		}
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
			term := highlightQuery(r.Term, c.Query, open, close)
			tags := ""
			if len(r.Tags) > 0 {
				tags = " [" + strings.Join(r.Tags, ",") + "]"
			}
			translations := highlightQuery(strings.Join(r.Translations, ", "), c.Query, open, close)
			fmt.Printf("  %s%s  → %s\n", term, tags, translations)
			if r.Notes != "" {
				notes := highlightQuery(r.Notes, c.Query, open, close)
				fmt.Printf("         notes: %s\n", notes)
			}
		}
		fmt.Println()
	}

	return nil
}

// grepMatchStyle returns the ANSI escape sequence for match highlighting
// based on GREP_COLORS or GREP_COLOR environment variables, falling back
// to bold red (\033[01;31m) if neither is set.
func grepMatchStyle() string {
	if gc := os.Getenv("GREP_COLORS"); gc != "" {
		for _, part := range strings.Split(gc, ":") {
			if strings.HasPrefix(part, "mt=") || strings.HasPrefix(part, "ms=") {
				if v := part[3:]; v != "" {
					return "\033[" + v + "m"
				}
			}
		}
	}
	if gc := os.Getenv("GREP_COLOR"); gc != "" {
		return "\033[" + gc + "m"
	}
	return "\033[01;31m"
}

// highlightQuery wraps all case-insensitive occurrences of query in text
// with the open and close ANSI sequences. Returns text unchanged when
// query or text is empty, or when open is empty.
func highlightQuery(text, query, open, close string) string {
	if query == "" || text == "" || open == "" {
		return text
	}
	lower := strings.ToLower(text)
	q := strings.ToLower(query)
	var b strings.Builder
	pos := 0
	for {
		idx := strings.Index(lower[pos:], q)
		if idx == -1 {
			b.WriteString(text[pos:])
			break
		}
		if idx > 0 {
			b.WriteString(text[pos : pos+idx])
		}
		b.WriteString(open)
		b.WriteString(text[pos+idx : pos+idx+len(q)])
		b.WriteString(close)
		pos += idx + len(q)
	}
	return b.String()
}
