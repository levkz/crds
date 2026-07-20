package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"crds/internal/parser"
	"crds/internal/ui"
)

type DeckStore struct {
	dir string
}

func NewDeckStore(dir string) *DeckStore {
	return &DeckStore{dir: dir}
}

func (s *DeckStore) ListDecks() ([]string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read decks dir: %w", err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".yaml")
		names = append(names, name)
	}
	return names, nil
}

func (s *DeckStore) LoadDeck(name string) (ui.DeckData, error) {
	path := filepath.Join(s.dir, name+".yaml")
	if _, err := os.Stat(path); err != nil {
		return ui.DeckData{}, fmt.Errorf("deck %q not found at %s", name, path)
	}

	deck, err := parser.ParseFile(path)
	if err != nil {
		return ui.DeckData{}, fmt.Errorf("load deck %q: %w", name, err)
	}

	cards := make([]ui.CardData, len(deck.Entries))
	for i, entry := range deck.Entries {
		cards[i] = entryToCardData(entry)
	}

	return ui.DeckData{
		Name:  deck.Name,
		Cards: cards,
	}, nil
}
