-- +goose Up

ALTER TABLE reviews ADD COLUMN reverse INTEGER NOT NULL DEFAULT 0;

DROP TABLE IF EXISTS progress;

CREATE TABLE progress (
    deck_id TEXT NOT NULL,
    entry_id TEXT NOT NULL,
    reverse INTEGER NOT NULL DEFAULT 0,
    ease REAL NOT NULL DEFAULT 2.5,
    interval INTEGER NOT NULL DEFAULT 0,
    due DATETIME,
    correct INTEGER NOT NULL DEFAULT 0,
    incorrect INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (deck_id, entry_id, reverse)
);

CREATE INDEX idx_progress_due ON progress(due);

-- +goose Down

DROP TABLE IF EXISTS progress;

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

CREATE INDEX idx_progress_due ON progress(due);

DROP TABLE IF EXISTS reviews;

CREATE TABLE reviews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    deck_id TEXT NOT NULL,
    entry_id TEXT NOT NULL,
    grade INTEGER NOT NULL,
    reviewed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
