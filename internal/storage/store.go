package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sync"
	"time"

	"crds/internal/storage/db"
	"crds/internal/ui"

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

// DB returns the underlying *sql.DB for advanced use (e.g. goose migrations outside of NewStore).
func (s *Store) DB() *sql.DB {
	return s.conn
}
