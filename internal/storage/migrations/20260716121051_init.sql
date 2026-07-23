-- +goose Up

CREATE TABLE sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at DATETIME,
    reviewed INTEGER NOT NULL DEFAULT 0,
    correct INTEGER NOT NULL DEFAULT 0,
    incorrect INTEGER NOT NULL DEFAULT 0,
    duration_ms INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE reviews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    deck_id TEXT NOT NULL,
    entry_id TEXT NOT NULL,
    grade INTEGER NOT NULL,
    reviewed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE typing_details (
    review_id INTEGER PRIMARY KEY REFERENCES reviews(id) ON DELETE CASCADE,
    user_input TEXT NOT NULL,
    correct_answer TEXT NOT NULL,
    similarity REAL NOT NULL
);

CREATE TABLE progress (
    deck_id TEXT NOT NULL,
    entry_id TEXT NOT NULL,
    ease REAL NOT NULL DEFAULT 2.5,
    interval INTEGER NOT NULL DEFAULT 0,
    due DATETIME,
    correct INTEGER NOT NULL DEFAULT 0,
    incorrect INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (deck_id, entry_id)
);

CREATE INDEX idx_reviews_session ON reviews(session_id);
CREATE INDEX idx_reviews_entry ON reviews(entry_id);
CREATE INDEX idx_reviews_deck ON reviews(deck_id);
CREATE INDEX idx_reviews_date ON reviews(reviewed_at);
CREATE INDEX idx_progress_due ON progress(due);

-- +goose Down
DROP INDEX IF EXISTS idx_reviews_session;
DROP INDEX IF EXISTS idx_reviews_entry;
DROP INDEX IF EXISTS idx_reviews_deck;
DROP INDEX IF EXISTS idx_reviews_date;
DROP INDEX IF EXISTS idx_progress_due;
DROP TABLE IF EXISTS typing_details;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS progress;
DROP TABLE IF EXISTS sessions;
