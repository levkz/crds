package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"crds/internal/model"
	"crds/internal/parser"
	"crds/internal/storage/db"
	"crds/internal/ui"
)

// SyncDecks parses all .yaml deck files from deckDir and syncs them into
// the SQLite cache. Only files whose mtime has changed since the last sync
// are re-parsed.
func (s *Store) SyncDecks(deckDir string) error {
	entries, err := os.ReadDir(deckDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read deck dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		path := filepath.Join(deckDir, entry.Name())
		if err := s.syncDeck(path); err != nil {
			log.Printf("warning: skipping %q: %v", entry.Name(), err)
		}
	}

	// Deduplicate translations and examples that may have accumulated due to
	// foreign keys not being enforced in earlier versions.
	ctx := context.Background()
	_, _ = s.conn.ExecContext(ctx, "DELETE FROM translations WHERE id NOT IN (SELECT MIN(id) FROM translations GROUP BY entry_id, text)")
	_, _ = s.conn.ExecContext(ctx, "DELETE FROM examples WHERE id NOT IN (SELECT MIN(id) FROM examples GROUP BY entry_id, text, translation)")
	// entry_tags has a composite PK so duplicates cannot exist.

	return nil
}

func (s *Store) syncDeck(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	needsSync, err := s.deckNeedsSync(path, info.ModTime())
	if err != nil {
		return err
	}
	if !needsSync {
		return nil
	}

	deck, err := parser.ParseFile(path)
	if err != nil {
		return err
	}

	ctx := context.Background()

	if err := s.queries.UpsertDeck(ctx, db.UpsertDeckParams{
		ID:                  deck.ID,
		Name:                deck.Name,
		Language:            deck.Language,
		TranslationLanguage: deck.TranslationLanguage,
	}); err != nil {
		return fmt.Errorf("upsert deck: %w", err)
	}

	// Explicitly delete child rows first as a safety net in case
	// PRAGMA foreign_keys was not honored by the driver.
	if _, err := s.conn.ExecContext(ctx, "DELETE FROM translations WHERE entry_id IN (SELECT id FROM entries WHERE deck_id = ?)", deck.ID); err != nil {
		return fmt.Errorf("delete translations by deck: %w", err)
	}
	if _, err := s.conn.ExecContext(ctx, "DELETE FROM examples WHERE entry_id IN (SELECT id FROM entries WHERE deck_id = ?)", deck.ID); err != nil {
		return fmt.Errorf("delete examples by deck: %w", err)
	}
	if _, err := s.conn.ExecContext(ctx, "DELETE FROM entry_tags WHERE entry_id IN (SELECT id FROM entries WHERE deck_id = ?)", deck.ID); err != nil {
		return fmt.Errorf("delete entry tags by deck: %w", err)
	}

	if err := s.queries.DeleteEntriesByDeck(ctx, deck.ID); err != nil {
		return fmt.Errorf("delete entries: %w", err)
	}

	for i, entry := range deck.Entries {
		notes := entry.Notes
		pos := int64(i)

		if err := s.queries.UpsertEntry(ctx, db.UpsertEntryParams{
			ID:       entry.ID,
			DeckID:   deck.ID,
			Term:     entry.Term,
			Notes:    notes,
			Position: pos,
		}); err != nil {
			return fmt.Errorf("upsert entry %q: %w", entry.ID, err)
		}

		for j, t := range entry.Translations {
			if err := s.queries.InsertTranslation(ctx, db.InsertTranslationParams{
				EntryID:  entry.ID,
				Text:     t.Text,
				Position: int64(j),
			}); err != nil {
				return fmt.Errorf("insert translation: %w", err)
			}
		}

		for j, ex := range entry.Examples {
			if err := s.queries.InsertExample(ctx, db.InsertExampleParams{
				EntryID:     entry.ID,
				Text:        ex.Text,
				Translation: ex.Translation,
				Position:    int64(j),
			}); err != nil {
				return fmt.Errorf("insert example: %w", err)
			}
		}

		for _, tag := range entry.Tags {
			if err := s.queries.InsertEntryTag(ctx, db.InsertEntryTagParams{
				EntryID: entry.ID,
				Tag:     tag,
			}); err != nil {
				return fmt.Errorf("insert tag: %w", err)
			}
		}
	}

	if err := s.queries.UpsertSyncState(ctx, db.UpsertSyncStateParams{
		Path:         path,
		LastModified: info.ModTime(),
	}); err != nil {
		return fmt.Errorf("upsert sync state: %w", err)
	}

	return nil
}

func (s *Store) deckNeedsSync(path string, mtime time.Time) (bool, error) {
	stored, err := s.queries.GetLastModified(context.Background(), path)
	if err != nil {
		// No sync state recorded yet — needs sync
		return true, nil
	}
	return mtime.After(stored), nil
}

// DeckSummary holds summary information for a listed deck.
type DeckSummary struct {
	ID                  string
	Name                string
	Language            string
	TranslationLanguage string
	EntryCount          int64
}

// ListDecksWithStats returns all cached decks with entry counts.
func (s *Store) ListDecksWithStats() ([]DeckSummary, error) {
	rows, err := s.queries.ListDecksWithStats(context.Background())
	if err != nil {
		return nil, err
	}
	stats := make([]DeckSummary, len(rows))
	for i, r := range rows {
		stats[i] = DeckSummary{
			ID:                  r.ID,
			Name:                r.Name,
			Language:            r.Language,
			TranslationLanguage: r.TranslationLanguage,
			EntryCount:          r.EntryCount,
		}
	}
	return stats, nil
}

// ListDecks returns the deck IDs of all cached decks.
func (s *Store) ListDecks() ([]string, error) {
	rows, err := s.queries.ListDeckNames(context.Background())
	if err != nil {
		return nil, err
	}
	names := make([]string, len(rows))
	for i, r := range rows {
		names[i] = r.ID
	}
	return names, nil
}

// LoadDeck returns the full deck data for the given deck ID.
func (s *Store) LoadDeck(id string) (ui.DeckData, error) {
	ctx := context.Background()

	d, err := s.queries.GetDeck(ctx, id)
	if err != nil {
		return ui.DeckData{}, fmt.Errorf("get deck %q: %w", id, err)
	}

	entries, err := s.queries.ListEntriesByDeck(ctx, id)
	if err != nil {
		return ui.DeckData{}, fmt.Errorf("list entries: %w", err)
	}

	cards := make([]ui.CardData, len(entries))
	for i, entry := range entries {
		cards[i] = s.buildCard(ctx, entry)
	}

	return ui.DeckData{
		Name:  d.Name,
		Cards: cards,
	}, nil
}

func (s *Store) buildCard(ctx context.Context, entry db.Entry) ui.CardData {
	texts, _ := s.queries.GetTranslationsByEntry(ctx, entry.ID)

	var variants []string
	for _, t := range texts {
		variants = append(variants, model.ExpandText(t)...)
	}

	return ui.CardData{
		ID:       entry.ID,
		Front:    entry.Term,
		Back:     texts,
		Variants: variants,
		Notes:    entry.Notes,
	}
}

// SearchResult holds a matched entry across a search.
type SearchResult struct {
	DeckID       string
	DeckName     string
	EntryID      string
	Term         string
	Translations []string
	Tags         []string
	Notes        string
}

// Search searches entries across the given decks (or all if nil/empty).
// Query is lowercased for case-insensitive matching. Tags are AND logic.
func (s *Store) Search(ctx context.Context, query string, deckIDs, tags []string) ([]SearchResult, error) {
	if len(deckIDs) == 0 {
		names, err := s.queries.ListDeckNames(ctx)
		if err != nil {
			return nil, fmt.Errorf("search: list decks: %w", err)
		}
		for _, n := range names {
			deckIDs = append(deckIDs, n.ID)
		}
	}

	query = strings.ToLower(query)
	var results []SearchResult

	for _, dID := range deckIDs {
		d, err := s.queries.GetDeck(ctx, dID)
		if err != nil {
			continue
		}

		entries, err := s.queries.ListEntriesByDeck(ctx, dID)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if query != "" && !matchesSearch(entry, query, ctx, s) {
				continue
			}

			etags, _ := s.queries.GetTagsByEntry(ctx, entry.ID)
			if !tagsMatch(etags, tags) {
				continue
			}

			translations, _ := s.queries.GetTranslationsByEntry(ctx, entry.ID)

			results = append(results, SearchResult{
				DeckID:       dID,
				DeckName:     d.Name,
				EntryID:      entry.ID,
				Term:         entry.Term,
				Translations: translations,
				Tags:         etags,
				Notes:        entry.Notes,
			})
		}
	}

	return results, nil
}

func matchesSearch(entry db.Entry, query string, ctx context.Context, s *Store) bool {
	if strings.Contains(strings.ToLower(entry.Term), query) {
		return true
	}
	if strings.Contains(strings.ToLower(entry.Notes), query) {
		return true
	}
	texts, _ := s.queries.GetTranslationsByEntry(ctx, entry.ID)
	for _, t := range texts {
		if strings.Contains(strings.ToLower(t), query) {
			return true
		}
	}
	return false
}

func tagsMatch(entryTags, filterTags []string) bool {
	if len(filterTags) == 0 {
		return true
	}
	set := make(map[string]bool, len(entryTags))
	for _, t := range entryTags {
		set[t] = true
	}
	for _, t := range filterTags {
		if !set[t] {
			return false
		}
	}
	return true
}

// EnsureStoreHasDecks syncs decks on first use if not yet synced.
func (s *Store) EnsureStoreHasDecks(deckDir string) error {
	rows, err := s.queries.ListDeckNames(context.Background())
	if err != nil {
		return err
	}
	if len(rows) > 0 {
		return nil
	}
	return s.SyncDecks(deckDir)
}
