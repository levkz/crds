package parser

import (
	"fmt"

	"crds/internal/model"
)

func Validate(deck *model.Deck) error {
	if deck.ID == "" {
		return fmt.Errorf("deck id is required")
	}

	if deck.Name == "" {
		return fmt.Errorf("deck name is required")
	}

	if deck.Language == "" {
		return fmt.Errorf("deck language is required")
	}

	ids := make(map[string]struct{})

	for i, entry := range deck.Entries {
		if entry.ID == "" {
			return fmt.Errorf("entry %d: missing id!\nrun `crds sync` to generate missing ids.", i+1)
		}

		if _, exists := ids[entry.ID]; exists {
			return fmt.Errorf("duplicate entry id: %q", entry.ID)
		}

		ids[entry.ID] = struct{}{}

		if entry.Term == "" {
			return fmt.Errorf("entry %q: missing term", entry.ID)
		}

		if len(entry.Translations) == 0 {
			return fmt.Errorf("entry %q: no translations", entry.Term)
		}
	}

	return nil
}
