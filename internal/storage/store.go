package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"crds/internal/model"
	"crds/internal/parser"
	"crds/internal/storage/db"
	"crds/internal/ui"

	"go.yaml.in/yaml/v3"
	_ "modernc.org/sqlite"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Store is the SQLite-backed persistence layer. It implements
// ProgressRecorder, StatsProvider, and provides session management.
type Store struct {
	queries *db.Queries
	conn    *sql.DB
	dbPath  string

	mu              sync.Mutex
	currentSession  int64 // 0 = no active session
	currentSessionAt time.Time
}

// NewStore opens or creates the SQLite database at dbPath, runs pending
// migrations, and returns a ready Store.
func NewStore(dbPath string) (*Store, error) {
	conn, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	// Run embedded migrations
	goose.SetBaseFS(migrationsFS)
	goose.SetDialect("sqlite")
	if err := goose.Up(conn, "migrations"); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}

	return &Store{
		queries: db.New(conn),
		conn:    conn,
		dbPath:  dbPath,
	}, nil
}

// Close shuts down the database connection.
func (s *Store) Close() error {
	return s.conn.Close()
}

// --- Session management ---

// EnsureSession returns an active session ID, creating one if needed.
func (s *Store) EnsureSession() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentSession != 0 {
		return s.currentSession, nil
	}

	session, err := s.queries.CreateSession(context.Background())
	if err != nil {
		return 0, fmt.Errorf("create session: %w", err)
	}
	s.currentSession = session.ID
	s.currentSessionAt = session.StartedAt
	return s.currentSession, nil
}

// ResetSession closes the current session (if any) and clears it so the
// next call to RecordAnswer will start a fresh session.
func (s *Store) ResetSession() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentSession == 0 {
		return nil
	}

	// Count reviews for this session
	reviews, err := s.queries.GetReviewsBySession(context.Background(), s.currentSession)
	if err == nil && len(reviews) > 0 {
		reviewed := int64(len(reviews))
		var correct int64
		for _, r := range reviews {
			if r.Grade >= 3 {
				correct++
			}
		}
		durationMs := time.Since(s.currentSessionAt).Milliseconds()
		_ = s.queries.FinishSession(context.Background(), db.FinishSessionParams{
			ID:         s.currentSession,
			Reviewed:   reviewed,
			Correct:    correct,
			Incorrect:  reviewed - correct,
			DurationMs: durationMs,
		})
	}

	s.currentSession = 0
	return nil
}

// --- RecordAnswer (ProgressRecorder interface) ---

// RecordAnswer persists a single quiz answer. An implicit session is created
// on the first call and reused until ResetSession is called.
func (s *Store) RecordAnswer(deckID, cardID string, grade int, reverse bool) error {
	sessionID, err := s.EnsureSession()
	if err != nil {
		return err
	}
	reverseInt := int64(0)
	if reverse {
		reverseInt = 1
	}
	_, err = s.queries.CreateReview(context.Background(), db.CreateReviewParams{
		SessionID: sessionID,
		DeckID:    deckID,
		EntryID:   cardID,
		Grade:     int64(grade),
		Reverse:   reverseInt,
	})
	return err
}

// RecordAnswerFull records a quiz answer with all available metadata.
func (s *Store) RecordAnswerFull(sessionID int64, deckID, entryID string, grade int, reverse bool, userInput, correctAnswer string, similarity float64) (int64, error) {
	if sessionID == 0 {
		var err error
		sessionID, err = s.EnsureSession()
		if err != nil {
			return 0, err
		}
	}

	reverseInt := int64(0)
	if reverse {
		reverseInt = 1
	}

	review, err := s.queries.CreateReview(context.Background(), db.CreateReviewParams{
		SessionID: sessionID,
		DeckID:    deckID,
		EntryID:   entryID,
		Grade:     int64(grade),
		Reverse:   reverseInt,
	})
	if err != nil {
		return 0, err
	}

	if userInput != "" || correctAnswer != "" {
		if err := s.queries.CreateTypingDetail(context.Background(), db.CreateTypingDetailParams{
			ReviewID:      review.ID,
			UserInput:     userInput,
			CorrectAnswer: correctAnswer,
			Similarity:    similarity,
		}); err != nil {
			return 0, err
		}
	}

	return review.ID, nil
}

// --- Stats (StatsProvider interface) ---

// Stats returns aggregate learning statistics for today.
func (s *Store) Stats() ui.Stats {
	row, err := s.queries.GetTodayStats(context.Background())
	if err != nil {
		return ui.Stats{}
	}

	var accuracy float64
	if row.TotalReviews > 0 {
		accuracy = float64(row.CorrectReviews) / float64(row.TotalReviews) * 100
	}

	// Total unique cards ever reviewed
	// For simplicity, return today's reviewed count as total cards
	return ui.Stats{
		ReviewedToday: int(row.TotalReviews),
		Accuracy:      accuracy,
		TotalCards:    int(row.TotalReviews),
	}
}

// --- Convenience queries ---

// GetReviewsByEntry returns the last n reviews for a given entry.
func (s *Store) GetReviewsByEntry(entryID string, limit int) ([]db.Review, error) {
	return s.queries.GetReviewsByEntry(context.Background(), db.GetReviewsByEntryParams{
		EntryID: entryID,
		Limit:   int64(limit),
	})
}

// GetWeakTypingEntries returns the weakest typed answers for a deck.
func (s *Store) GetWeakTypingEntries(deckID string, limit int) ([]db.GetWeakTypingEntriesRow, error) {
	return s.queries.GetWeakTypingEntries(context.Background(), db.GetWeakTypingEntriesParams{
		DeckID: deckID,
		Limit:  int64(limit),
	})
}

// --- Deck operations ---

// ImportDeck parses a YAML file, copies it into the deck directory, and syncs
// it into the SQLite cache. Returns an error if the deck ID already exists in
// the database or if a file with the deck ID already exists in the directory.
func (s *Store) ImportDeck(srcPath, deckDir string) error {
	ctx := context.Background()

	deck, err := parser.ParseFile(srcPath)
	if err != nil {
		return fmt.Errorf("import: parse %s: %w", srcPath, err)
	}

	_, err = s.queries.GetDeck(ctx, deck.ID)
	if err == nil {
		return fmt.Errorf("import: deck %q already exists in database", deck.ID)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("import: check deck %q: %w", deck.ID, err)
	}

	dstPath := filepath.Join(deckDir, deck.ID+".yaml")
	if _, err := os.Stat(dstPath); err == nil {
		return fmt.Errorf("import: file %q already exists", dstPath)
	}

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("import: read %s: %w", srcPath, err)
	}
	if err := os.WriteFile(dstPath, data, 0644); err != nil {
		return fmt.Errorf("import: write %s: %w", dstPath, err)
	}

	if err := s.syncDeck(dstPath); err != nil {
		os.Remove(dstPath)
		return fmt.Errorf("import: sync %s: %w", dstPath, err)
	}

	return nil
}

// ExportDeck copies the source YAML file for a deck to the destination path.
// Returns an error if the deck or source file does not exist.
func (s *Store) ExportDeck(deckID, dstPath, deckDir string) error {
	_, err := s.queries.GetDeck(context.Background(), deckID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("export: deck %q not found", deckID)
		}
		return fmt.Errorf("export: get deck %q: %w", deckID, err)
	}

	srcPath := filepath.Join(deckDir, deckID+".yaml")
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("export: read %s: %w", srcPath, err)
	}
	if err := os.WriteFile(dstPath, data, 0644); err != nil {
		return fmt.Errorf("export: write %s: %w", dstPath, err)
	}

	return nil
}

// ExportDeckFromCache reconstructs a deck from the SQLite cache and writes it
// as canonical YAML to the destination path. Preserves no comments or original
// formatting. Works even if the source YAML file has been deleted.
func (s *Store) ExportDeckFromCache(deckID, dstPath string) error {
	ctx := context.Background()

	d, err := s.queries.GetDeck(ctx, deckID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("export-cache: deck %q not found", deckID)
		}
		return fmt.Errorf("export-cache: get deck %q: %w", deckID, err)
	}

	entries, err := s.queries.ListEntriesByDeck(ctx, deckID)
	if err != nil {
		return fmt.Errorf("export-cache: list entries: %w", err)
	}

	deck := &model.Deck{
		ID:                  d.ID,
		Name:                d.Name,
		Language:            d.Language,
		TranslationLanguage: d.TranslationLanguage,
	}

	for _, entry := range entries {
		e := model.Entry{
			ID:    entry.ID,
			Term:  entry.Term,
			Notes: entry.Notes,
		}

		texts, _ := s.queries.GetTranslationsByEntry(ctx, entry.ID)
		for _, t := range texts {
			e.Translations = append(e.Translations, model.Translation{Text: t})
		}

		examples, _ := s.queries.GetExamplesByEntry(ctx, entry.ID)
		for _, ex := range examples {
			e.Examples = append(e.Examples, model.Example{
				Text:        ex.Text,
				Translation: ex.Translation,
			})
		}

		tags, _ := s.queries.GetTagsByEntry(ctx, entry.ID)
		e.Tags = tags

		deck.Entries = append(deck.Entries, e)
	}

	data, err := yaml.Marshal(deck)
	if err != nil {
		return fmt.Errorf("export-cache: marshal: %w", err)
	}

	if err := os.WriteFile(dstPath, data, 0644); err != nil {
		return fmt.Errorf("export-cache: write %s: %w", dstPath, err)
	}

	return nil
}

// RenameDeck changes the name of a deck in both the source YAML file and the
// SQLite cache.
func (s *Store) RenameDeck(deckID, newName, deckDir string) error {
	yamlPath := filepath.Join(deckDir, deckID+".yaml")

	deck, err := parser.ParseFile(yamlPath)
	if err != nil {
		return fmt.Errorf("rename: parse %s: %w", yamlPath, err)
	}

	deck.Name = newName

	data, err := yaml.Marshal(deck)
	if err != nil {
		return fmt.Errorf("rename: marshal: %w", err)
	}
	if err := os.WriteFile(yamlPath, data, 0644); err != nil {
		return fmt.Errorf("rename: write %s: %w", yamlPath, err)
	}

	_, err = s.conn.ExecContext(context.Background(),
		"UPDATE decks SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		newName, deckID)
	if err != nil {
		return fmt.Errorf("rename: update db: %w", err)
	}

	return nil
}

// ChangeDeckID changes the ID of a deck in the source YAML file, SQLite cache,
// and all referencing tables (entries, progress, reviews, sync_state).
func (s *Store) ChangeDeckID(deckID, newID, deckDir string) error {
	ctx := context.Background()

	_, err := s.queries.GetDeck(ctx, newID)
	if err == nil {
		return fmt.Errorf("change-id: deck %q already exists", newID)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("change-id: check deck %q: %w", newID, err)
	}

	yamlPath := filepath.Join(deckDir, deckID+".yaml")
	deck, err := parser.ParseFile(yamlPath)
	if err != nil {
		return fmt.Errorf("change-id: parse %s: %w", yamlPath, err)
	}

	deck.ID = newID

	newYamlPath := filepath.Join(deckDir, newID+".yaml")
	data, err := yaml.Marshal(deck)
	if err != nil {
		return fmt.Errorf("change-id: marshal: %w", err)
	}
	if err := os.WriteFile(newYamlPath, data, 0644); err != nil {
		return fmt.Errorf("change-id: write %s: %w", newYamlPath, err)
	}

	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		os.Remove(newYamlPath)
		return fmt.Errorf("change-id: begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO decks (id, name, language, translation_language, created_at, updated_at)
		 SELECT ?, name, language, translation_language, created_at, CURRENT_TIMESTAMP
		 FROM decks WHERE id = ?`,
		newID, deckID)
	if err != nil {
		return fmt.Errorf("change-id: insert deck: %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE entries SET deck_id = ? WHERE deck_id = ?", newID, deckID)
	if err != nil {
		return fmt.Errorf("change-id: update entries: %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE progress SET deck_id = ? WHERE deck_id = ?", newID, deckID)
	if err != nil {
		return fmt.Errorf("change-id: update progress: %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE reviews SET deck_id = ? WHERE deck_id = ?", newID, deckID)
	if err != nil {
		return fmt.Errorf("change-id: update reviews: %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE sync_state SET path = ? WHERE path = ?", newYamlPath, yamlPath)
	if err != nil {
		return fmt.Errorf("change-id: update sync_state: %w", err)
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM decks WHERE id = ?", deckID)
	if err != nil {
		return fmt.Errorf("change-id: delete old deck: %w", err)
	}

	if err := tx.Commit(); err != nil {
		os.Remove(newYamlPath)
		return fmt.Errorf("change-id: commit: %w", err)
	}

	if err := os.Remove(yamlPath); err != nil {
		return fmt.Errorf("change-id: remove old file %s: %w", yamlPath, err)
	}

	return nil
}

// DeleteDeck removes a deck from the database and filesystem. It deletes
// associated progress, reviews, and the source YAML file. The deck row is
// removed via ON DELETE CASCADE which cleans up entries, translations,
// examples, and entry_tags automatically.
func (s *Store) DeleteDeck(deckID, deckDir string) error {
	ctx := context.Background()

	_, err := s.queries.GetDeck(ctx, deckID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("delete: deck %q not found", deckID)
		}
		return fmt.Errorf("delete: check deck %q: %w", deckID, err)
	}

	yamlPath := filepath.Join(deckDir, deckID+".yaml")

	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("delete: begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "DELETE FROM progress WHERE deck_id = ?", deckID)
	if err != nil {
		return fmt.Errorf("delete: progress: %w", err)
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM reviews WHERE deck_id = ?", deckID)
	if err != nil {
		return fmt.Errorf("delete: reviews: %w", err)
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM sync_state WHERE path = ?", yamlPath)
	if err != nil {
		return fmt.Errorf("delete: sync_state: %w", err)
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM decks WHERE id = ?", deckID)
	if err != nil {
		return fmt.Errorf("delete: deck: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("delete: commit: %w", err)
	}

	if err := os.Remove(yamlPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete: remove file %s: %w", yamlPath, err)
	}

	return nil
}

// --- Entry operations ---

// AddEntry appends a new entry to a deck's YAML file and syncs it to the DB.
// Returns an error if the entry ID already exists in the deck.
func (s *Store) AddEntry(deckID string, entry model.Entry, deckDir string) error {
	yamlPath := filepath.Join(deckDir, deckID+".yaml")

	deck, err := parser.ParseFile(yamlPath)
	if err != nil {
		return fmt.Errorf("add-entry: parse %s: %w", yamlPath, err)
	}

	for _, e := range deck.Entries {
		if e.ID == entry.ID {
			return fmt.Errorf("add-entry: entry %q already exists in deck %q", entry.ID, deckID)
		}
	}

	deck.Entries = append(deck.Entries, entry)

	data, err := yaml.Marshal(deck)
	if err != nil {
		return fmt.Errorf("add-entry: marshal: %w", err)
	}
	if err := os.WriteFile(yamlPath, data, 0644); err != nil {
		return fmt.Errorf("add-entry: write: %w", err)
	}

	if err := s.syncDeck(yamlPath); err != nil {
		return fmt.Errorf("add-entry: sync: %w", err)
	}

	return nil
}

// UpdateEntry replaces an existing entry's fields (same ID) in the YAML file
// and syncs to the DB. The entry ID must match the existing entry's ID.
func (s *Store) UpdateEntry(deckID, entryID string, entry model.Entry, deckDir string) error {
	yamlPath := filepath.Join(deckDir, deckID+".yaml")

	deck, err := parser.ParseFile(yamlPath)
	if err != nil {
		return fmt.Errorf("update-entry: parse %s: %w", yamlPath, err)
	}

	found := false
	for i, e := range deck.Entries {
		if e.ID == entryID {
			entry.ID = entryID
			deck.Entries[i] = entry
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("update-entry: entry %q not found in deck %q", entryID, deckID)
	}

	data, err := yaml.Marshal(deck)
	if err != nil {
		return fmt.Errorf("update-entry: marshal: %w", err)
	}
	if err := os.WriteFile(yamlPath, data, 0644); err != nil {
		return fmt.Errorf("update-entry: write: %w", err)
	}

	if err := s.syncDeck(yamlPath); err != nil {
		return fmt.Errorf("update-entry: sync: %w", err)
	}

	return nil
}

// ReplaceEntryID changes an entry's ID in the YAML file and migrates progress
// and review records from the old ID to the new ID so statistics are preserved.
func (s *Store) ReplaceEntryID(deckID, oldID, newID string, deckDir string) error {
	yamlPath := filepath.Join(deckDir, deckID+".yaml")

	deck, err := parser.ParseFile(yamlPath)
	if err != nil {
		return fmt.Errorf("replace-id: parse %s: %w", yamlPath, err)
	}

	for _, e := range deck.Entries {
		if e.ID == newID {
			return fmt.Errorf("replace-id: entry %q already exists in deck %q", newID, deckID)
		}
	}

	found := false
	for i, e := range deck.Entries {
		if e.ID == oldID {
			deck.Entries[i].ID = newID
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("replace-id: entry %q not found in deck %q", oldID, deckID)
	}

	data, err := yaml.Marshal(deck)
	if err != nil {
		return fmt.Errorf("replace-id: marshal: %w", err)
	}
	if err := os.WriteFile(yamlPath, data, 0644); err != nil {
		return fmt.Errorf("replace-id: write: %w", err)
	}

	ctx := context.Background()
	_, err = s.conn.ExecContext(ctx, "UPDATE progress SET entry_id = ? WHERE deck_id = ? AND entry_id = ?", newID, deckID, oldID)
	if err != nil {
		return fmt.Errorf("replace-id: progress: %w", err)
	}
	_, err = s.conn.ExecContext(ctx, "UPDATE reviews SET entry_id = ? WHERE deck_id = ? AND entry_id = ?", newID, deckID, oldID)
	if err != nil {
		return fmt.Errorf("replace-id: reviews: %w", err)
	}

	if err := s.syncDeck(yamlPath); err != nil {
		return fmt.Errorf("replace-id: sync: %w", err)
	}

	return nil
}

// RemoveEntry deletes an entry from the YAML file, syncs to the DB, and cleans
// up associated progress and review records.
func (s *Store) RemoveEntry(deckID, entryID, deckDir string) error {
	yamlPath := filepath.Join(deckDir, deckID+".yaml")

	deck, err := parser.ParseFile(yamlPath)
	if err != nil {
		return fmt.Errorf("remove-entry: parse %s: %w", yamlPath, err)
	}

	found := false
	for i, e := range deck.Entries {
		if e.ID == entryID {
			deck.Entries = append(deck.Entries[:i], deck.Entries[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("remove-entry: entry %q not found in deck %q", entryID, deckID)
	}

	data, err := yaml.Marshal(deck)
	if err != nil {
		return fmt.Errorf("remove-entry: marshal: %w", err)
	}
	if err := os.WriteFile(yamlPath, data, 0644); err != nil {
		return fmt.Errorf("remove-entry: write: %w", err)
	}

	if err := s.syncDeck(yamlPath); err != nil {
		return fmt.Errorf("remove-entry: sync: %w", err)
	}

	ctx := context.Background()
	_, err = s.conn.ExecContext(ctx, "DELETE FROM progress WHERE deck_id = ? AND entry_id = ?", deckID, entryID)
	if err != nil {
		return fmt.Errorf("remove-entry: progress: %w", err)
	}
	_, err = s.conn.ExecContext(ctx, "DELETE FROM reviews WHERE deck_id = ? AND entry_id = ?", deckID, entryID)
	if err != nil {
		return fmt.Errorf("remove-entry: reviews: %w", err)
	}

	return nil
}

// AddTagsToEntry adds tags to an entry in the YAML file and syncs to the DB.
// Duplicate tags are ignored.
func (s *Store) AddTagsToEntry(deckID, entryID string, tags []string, deckDir string) error {
	yamlPath := filepath.Join(deckDir, deckID+".yaml")

	deck, err := parser.ParseFile(yamlPath)
	if err != nil {
		return fmt.Errorf("add-tags: parse %s: %w", yamlPath, err)
	}

	var entry *model.Entry
	for i := range deck.Entries {
		if deck.Entries[i].ID == entryID {
			entry = &deck.Entries[i]
			break
		}
	}
	if entry == nil {
		return fmt.Errorf("add-tags: entry %q not found in deck %q", entryID, deckID)
	}

	existing := make(map[string]bool, len(entry.Tags))
	for _, t := range entry.Tags {
		existing[t] = true
	}
	for _, t := range tags {
		if !existing[t] {
			entry.Tags = append(entry.Tags, t)
			existing[t] = true
		}
	}

	data, err := yaml.Marshal(deck)
	if err != nil {
		return fmt.Errorf("add-tags: marshal: %w", err)
	}
	if err := os.WriteFile(yamlPath, data, 0644); err != nil {
		return fmt.Errorf("add-tags: write: %w", err)
	}

	return s.syncDeck(yamlPath)
}

// RemoveTagsFromEntry removes specific tags from an entry in the YAML file
// and syncs to the DB. Tags that don't exist are silently ignored.
func (s *Store) RemoveTagsFromEntry(deckID, entryID string, tags []string, deckDir string) error {
	yamlPath := filepath.Join(deckDir, deckID+".yaml")

	deck, err := parser.ParseFile(yamlPath)
	if err != nil {
		return fmt.Errorf("remove-tags: parse %s: %w", yamlPath, err)
	}

	var entry *model.Entry
	for i := range deck.Entries {
		if deck.Entries[i].ID == entryID {
			entry = &deck.Entries[i]
			break
		}
	}
	if entry == nil {
		return fmt.Errorf("remove-tags: entry %q not found in deck %q", entryID, deckID)
	}

	remove := make(map[string]bool, len(tags))
	for _, t := range tags {
		remove[t] = true
	}
	filtered := make([]string, 0, len(entry.Tags))
	for _, t := range entry.Tags {
		if !remove[t] {
			filtered = append(filtered, t)
		}
	}
	entry.Tags = filtered

	data, err := yaml.Marshal(deck)
	if err != nil {
		return fmt.Errorf("remove-tags: marshal: %w", err)
	}
	if err := os.WriteFile(yamlPath, data, 0644); err != nil {
		return fmt.Errorf("remove-tags: write: %w", err)
	}

	return s.syncDeck(yamlPath)
}

// GetTagsByEntry returns the tags for a given entry, read from the SQLite cache.
func (s *Store) GetTagsByEntry(entryID string) ([]string, error) {
	return s.queries.GetTagsByEntry(context.Background(), entryID)
}

// DB returns the underlying *sql.DB for advanced use (e.g. goose migrations outside of NewStore).
func (s *Store) DB() *sql.DB {
	return s.conn
}
