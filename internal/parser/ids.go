package parser

import (
	"fmt"
	"strings"
	"unicode"

	"crds/internal/model"
)

func assignIDs(deck *model.Deck) {
	used := make(map[string]bool)
	for _, entry := range deck.Entries {
		if entry.ID != "" {
			used[entry.ID] = true
		}
	}
	for i := range deck.Entries {
		entry := &deck.Entries[i]
		if entry.ID != "" {
			continue
		}
		base := simplifyID(entry.Term)
		if base == "" {
			base = "entry"
		}
		id := base
		for suffix := 2; used[id]; suffix++ {
			id = fmt.Sprintf("%s_%d", base, suffix)
		}
		used[id] = true
		entry.ID = id
	}
}

func simplifyID(term string) string {
	variants := model.ExpandText(term)
	best := variants[0]
	for _, v := range variants[1:] {
		if len(v) < len(best) {
			best = v
		}
	}
	return sanitizeID(best)
}

func sanitizeID(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r == ' ' || r == '\'' || r == '-':
			b.WriteRune('_')
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		}
	}
	return b.String()
}
