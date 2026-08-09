package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"crds/internal/model"
	"crds/internal/parser"
	"crds/internal/stats"
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

// Summary returns aggregate stats across all decks (stats.Provider).
func (s *Store) Summary() (stats.Summary, error) {
	row, err := s.queries.GetTodayStats(context.Background())
	if err != nil {
		return stats.Summary{}, err
	}

	var accuracy float64
	if row.TotalReviews > 0 {
		accuracy = float64(row.CorrectReviews) / float64(row.TotalReviews) * 100
	}

	mastered := s.masteredCount()

	return stats.Summary{
		ReviewedToday: int(row.TotalReviews),
		Accuracy:      accuracy,
		TotalCards:    int(row.TotalReviews),
		Mastered:      mastered,
		DueToday:      0,
	}, nil
}

func (s *Store) masteredCount() int {
	ctx := context.Background()
	allProgress, err := s.queries.GetAllProgress(ctx)
	if err != nil {
		return 0
	}
	var count int
	for _, p := range allProgress {
		correct := int(p.Correct)
		incorrect := int(p.Incorrect)
		if stats.Confidence(correct, incorrect) >= 0.8 {
			count++
		}
	}
	return count
}

// DeckSummary returns per-deck stats (stats.Provider).
func (s *Store) DeckSummary(deckID string) (stats.DeckStats, error) {
	ctx := context.Background()

	entryCount, err := s.queries.GetDeckEntryCount(ctx, deckID)
	if err != nil {
		return stats.DeckStats{}, fmt.Errorf("deck stats: get entry count: %w", err)
	}

	row, err := s.queries.GetTodayStatsByDeck(ctx, deckID)
	if err != nil {
		return stats.DeckStats{}, fmt.Errorf("deck stats: get today stats: %w", err)
	}

	var accuracy float64
	if row.TotalReviews > 0 {
		accuracy = float64(row.CorrectReviews) / float64(row.TotalReviews) * 100
	}

	progress, err := s.queries.GetDeckProgress(ctx, deckID)
	if err != nil {
		progress = nil
	}
	var totalConf float64
	for _, p := range progress {
		totalConf += stats.Confidence(int(p.Correct), int(p.Incorrect))
	}
	var avgConf float64
	if len(progress) > 0 {
		avgConf = totalConf / float64(len(progress))
	}

	return stats.DeckStats{
		TotalEntries:  int(entryCount),
		ReviewedToday: int(row.TotalReviews),
		Accuracy:      accuracy,
		AvgConfidence: avgConf,
	}, nil
}

// EntryProgress returns per-entry progress for a deck (stats.Provider).
func (s *Store) EntryProgress(deckID string) (map[string]stats.EntryProgress, error) {
	ctx := context.Background()
	rows, err := s.queries.GetDeckProgress(ctx, deckID)
	if err != nil {
		return nil, fmt.Errorf("entry progress: %w", err)
	}
	result := make(map[string]stats.EntryProgress, len(rows))
	for _, r := range rows {
		result[r.EntryID] = stats.EntryProgress{
			Correct:   int(r.Correct),
			Incorrect: int(r.Incorrect),
		}
	}
	return result, nil
}

// selectionDecks resolves a deck/tag selection to a concrete deck list.
// Rule: non-empty deckIDs win; otherwise decks carrying any selected tag;
// otherwise all decks (empty slice signals "no filter").
func (s *Store) selectionDecks(deckIDs, tags []string) ([]string, error) {
	if len(deckIDs) > 0 {
		return deckIDs, nil
	}
	if len(tags) > 0 {
		seen := make(map[string]bool)
		for _, t := range tags {
			rows, err := s.queries.ListDecksByTag(context.Background(), t)
			if err != nil {
				return nil, err
			}
			for _, d := range rows {
				seen[d] = true
			}
		}
		var out []string
		for d := range seen {
			out = append(out, d)
		}
		return out, nil
	}
	return nil, nil
}

// SelectionSummary returns aggregate stats for the deck/tag selection
// (see selectionDecks for the scope rules).
func (s *Store) SelectionSummary(deckIDs, tags []string) (stats.Summary, error) {
	ctx := context.Background()
	decks, err := s.selectionDecks(deckIDs, tags)
	if err != nil {
		return stats.Summary{}, err
	}

	var total, correct int64
	if len(decks) == 0 {
		row, err := s.queries.GetTodayStats(ctx)
		if err != nil {
			return stats.Summary{}, err
		}
		total, correct = row.TotalReviews, row.CorrectReviews
	} else {
		row, err := s.queries.GetTodayStatsByDecks(ctx, decks)
		if err != nil {
			return stats.Summary{}, err
		}
		total, correct = row.TotalReviews, row.CorrectReviews
	}

	var accuracy float64
	if total > 0 {
		accuracy = float64(correct) / float64(total) * 100
	}

	mastered := s.masteredCountIn(decks)

	days, err := s.selectionReviewDays(decks)
	if err != nil {
		return stats.Summary{}, err
	}

	return stats.Summary{
		ReviewedToday: int(total),
		Accuracy:      accuracy,
		TotalCards:    s.entryCountIn(decks),
		Mastered:      mastered,
		DueToday:      0,
		Streak:        stats.Streak(days),
	}, nil
}

// entryCountIn counts distinct entries across the given decks. Entry IDs are
// globally unique (PK), so summing per-deck counts is exact. An empty deck
// list means all decks.
func (s *Store) entryCountIn(decks []string) int {
	ctx := context.Background()
	if len(decks) == 0 {
		all, err := s.queries.GetAllEntries(ctx)
		if err != nil {
			return 0
		}
		return len(all)
	}
	total := 0
	for _, d := range decks {
		n, err := s.queries.GetDeckEntryCount(ctx, d)
		if err != nil {
			continue
		}
		total += int(n)
	}
	return total
}

// SelectionHistory returns daily review aggregates for the selection.
func (s *Store) SelectionHistory(deckIDs, tags []string) ([]stats.DayPoint, error) {
	decks, err := s.selectionDecks(deckIDs, tags)
	if err != nil {
		return nil, err
	}
	var out []stats.DayPoint
	if len(decks) == 0 {
		rows, err := s.queries.GetDailyStats(context.Background())
		if err != nil {
			return nil, err
		}
		out = make([]stats.DayPoint, 0, len(rows))
		for _, r := range rows {
			out = append(out, stats.DayPoint{
				Day:       r.Day,
				Correct:   int(r.CorrectReviews),
				Incorrect: int(r.TotalReviews - r.CorrectReviews),
			})
		}
		return out, nil
	}
	rows, err := s.queries.GetDailyStatsByDecks(context.Background(), decks)
	if err != nil {
		return nil, err
	}
	out = make([]stats.DayPoint, 0, len(rows))
	for _, r := range rows {
		out = append(out, stats.DayPoint{
			Day:       r.Day,
			Correct:   int(r.CorrectReviews),
			Incorrect: int(r.TotalReviews - r.CorrectReviews),
		})
	}
	return out, nil
}

// WordStats returns per-entry statistics for a single entry.
func (s *Store) WordStats(entryID string) (stats.WordStats, error) {
	row, err := s.queries.GetEntryStats(context.Background(), entryID)
	if err != nil {
		return stats.WordStats{}, err
	}
	ws := stats.WordStats{
		TotalReviews:  int(row.TotalReviews),
		ReviewedToday: int(row.ReviewedToday),
		Correct:       int(row.CorrectReviews),
		Incorrect:     int(row.IncorrectReviews),
	}
	if row.LastReviewed != "" {
		if t, err := parseSQLTime(row.LastReviewed); err == nil {
			ws.LastReviewed = &t
		}
	}
	return ws, nil
}

// WordHistory returns daily review aggregates for a single entry.
func (s *Store) WordHistory(entryID string) ([]stats.DayPoint, error) {
	rows, err := s.queries.GetEntryDailyStats(context.Background(), entryID)
	if err != nil {
		return nil, err
	}
	out := make([]stats.DayPoint, 0, len(rows))
	for _, r := range rows {
		out = append(out, stats.DayPoint{
			Day:       r.Day,
			Correct:   int(r.CorrectReviews),
			Incorrect: int(r.TotalReviews - r.CorrectReviews),
		})
	}
	return out, nil
}

func (s *Store) selectionReviewDays(decks []string) ([]time.Time, error) {
	var days []string
	var err error
	if len(decks) == 0 {
		days, err = s.queries.GetReviewDays(context.Background())
	} else {
		days, err = s.queries.GetReviewDaysByDecks(context.Background(), decks)
	}
	if err != nil {
		return nil, err
	}
	var out []time.Time
	for _, d := range days {
		if t, err := time.Parse("2006-01-02", d); err == nil {
			out = append(out, t)
		}
	}
	return out, nil
}

// masteredCountIn counts entries at/above the mastery threshold, restricted
// to the given decks (nil/empty = all decks).
func (s *Store) masteredCountIn(decks []string) int {
	ctx := context.Background()
	var all []db.GetDeckProgressRow
	for _, d := range decks {
		rows, err := s.queries.GetDeckProgress(ctx, d)
		if err != nil {
			continue
		}
		all = append(all, rows...)
	}
	if len(decks) == 0 {
		allProgress, err := s.queries.GetAllProgress(ctx)
		if err != nil {
			return 0
		}
		var count int
		for _, p := range allProgress {
			correct := int(p.Correct)
			incorrect := int(p.Incorrect)
			if stats.Confidence(correct, incorrect) >= 0.8 {
				count++
			}
		}
		return count
	}

	// Deduplicate entries appearing in multiple decks.
	seen := make(map[string]bool)
	count := 0
	for _, p := range all {
		correct := int(p.Correct)
		incorrect := int(p.Incorrect)
		if stats.Confidence(correct, incorrect) >= 0.8 && !seen[p.EntryID] {
			seen[p.EntryID] = true
			count++
		}
	}
	return count
}

func parseSQLTime(s string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("parse time %q", s)
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

// DeckStats returns aggregate statistics for a single deck.
type DeckStats struct {
	TotalEntries  int
	ReviewedToday int
	Accuracy      float64
}

// DeckStats returns per-deck learning statistics.
func (s *Store) DeckStats(deckID string) (DeckStats, error) {
	ctx := context.Background()

	entryCount, err := s.queries.GetDeckEntryCount(ctx, deckID)
	if err != nil {
		return DeckStats{}, fmt.Errorf("deck stats: get entry count: %w", err)
	}

	row, err := s.queries.GetTodayStatsByDeck(ctx, deckID)
	if err != nil {
		return DeckStats{}, fmt.Errorf("deck stats: get today stats: %w", err)
	}

	var accuracy float64
	if row.TotalReviews > 0 {
		accuracy = float64(row.CorrectReviews) / float64(row.TotalReviews) * 100
	}

	return DeckStats{
		TotalEntries:  int(entryCount),
		ReviewedToday: int(row.TotalReviews),
		Accuracy:      accuracy,
	}, nil
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

// ListAllTags returns every unique tag across all decks.
func (s *Store) ListAllTags() ([]string, error) {
	return s.queries.ListAllTags(context.Background())
}

// ListDeckTags returns all tags belonging to a single deck.
func (s *Store) ListDeckTags(deckID string) ([]string, error) {
	return s.queries.ListDeckTags(context.Background(), deckID)
}

// ListDecksByTag returns deck IDs that have the given tag.
func (s *Store) ListDecksByTag(tag string) ([]string, error) {
	return s.queries.ListDecksByTag(context.Background(), tag)
}

// FilterDecksByTags returns deck IDs that have ALL of the given tags (AND logic).
func (s *Store) FilterDecksByTags(tags []string) ([]string, error) {
	if len(tags) == 0 {
		return s.ListDecks()
	}

	ctx := context.Background()

	// Build placeholders for IN clause
	placeholders := make([]string, len(tags))
	args := make([]any, len(tags)+1)
	for i, tag := range tags {
		placeholders[i] = "?"
		args[i] = tag
	}
	args[len(tags)] = len(tags)

	query := fmt.Sprintf(
		`SELECT deck_id FROM deck_tags WHERE tag IN (%s) GROUP BY deck_id HAVING COUNT(DISTINCT tag) = ?`,
		strings.Join(placeholders, ","),
	)

	rows, err := s.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("filter decks by tags: %w", err)
	}
	defer rows.Close()

	var deckIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		deckIDs = append(deckIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return deckIDs, nil
}

// FilterTagsByDecks returns tags that belong to ALL of the given decks (intersection).
func (s *Store) FilterTagsByDecks(deckIDs []string) ([]string, error) {
	if len(deckIDs) == 0 {
		return s.ListAllTags()
	}

	ctx := context.Background()

	placeholders := make([]string, len(deckIDs))
	args := make([]any, len(deckIDs)+1)
	for i, id := range deckIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	args[len(deckIDs)] = len(deckIDs)

	query := fmt.Sprintf(
		`SELECT tag FROM deck_tags WHERE deck_id IN (%s) GROUP BY tag HAVING COUNT(DISTINCT deck_id) = ?`,
		strings.Join(placeholders, ","),
	)

	rows, err := s.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("filter tags by decks: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tags, nil
}

// DB returns the underlying *sql.DB for advanced use (e.g. goose migrations outside of NewStore).
func (s *Store) DB() *sql.DB {
	return s.conn
}
