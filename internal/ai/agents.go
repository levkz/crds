package ai

import (
	"context"
	"strings"

	"crds/internal/model"
)

// Interpret runs the interpreter agent: free-form text into minimal entries.
func Interpret(ctx context.Context, c Client, raw string, lc LanguageContext, msg string) ([]model.Entry, error) {
	system, user := InterpretMessages(raw, lc, msg)
	reply, err := c.Complete(ctx, system, user)
	if err != nil {
		return nil, err
	}
	return ParseEntries(reply)
}

// InterpretFull runs full-effort interpretation: free-form text into complete
// entries (examples, notes, deck-constrained tags) in a single call.
func InterpretFull(ctx context.Context, c Client, raw string, dc DeckContext, msg string) ([]model.Entry, error) {
	system, user := InterpretFullMessages(raw, dc, msg)
	reply, err := c.Complete(ctx, system, user)
	if err != nil {
		return nil, err
	}
	return ParseEntries(reply)
}

// Fill runs the filler agent: partial entries into completed entries.
func Fill(ctx context.Context, c Client, entries []model.Entry, dc DeckContext, msg string) ([]model.Entry, error) {
	system, user := FillMessages(entries, dc, msg)
	reply, err := c.Complete(ctx, system, user)
	if err != nil {
		return nil, err
	}
	return ParseEntries(reply)
}

// IsStructuredInput reports whether the text looks like YAML entries already
// rather than free-form words. Used by `crds ai add` to pick a path.
func IsStructuredInput(input string) bool {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return false
	}
	lower := strings.ToLower(trimmed)
	for _, marker := range []string{"- term:", "- id:", "term:", "id:", "entries:"} {
		if strings.HasPrefix(lower, marker) {
			return true
		}
	}
	return false
}