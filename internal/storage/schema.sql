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
    reverse INTEGER NOT NULL DEFAULT 0,
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
    reverse INTEGER NOT NULL DEFAULT 0,
    ease REAL NOT NULL DEFAULT 2.5,
    interval INTEGER NOT NULL DEFAULT 0,
    due DATETIME,
    correct INTEGER NOT NULL DEFAULT 0,
    incorrect INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (deck_id, entry_id, reverse)
);

CREATE INDEX idx_reviews_session ON reviews(session_id);
CREATE INDEX idx_reviews_entry ON reviews(entry_id);
CREATE INDEX idx_reviews_deck ON reviews(deck_id);
CREATE INDEX idx_reviews_date ON reviews(reviewed_at);
CREATE INDEX idx_progress_due ON progress(due);

CREATE TABLE decks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    language TEXT NOT NULL,
    translation_language TEXT NOT NULL,
    input_mappings TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE entries (
    id TEXT PRIMARY KEY,
    deck_id TEXT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    term TEXT NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    position INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE translations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entry_id TEXT NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    position INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE examples (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entry_id TEXT NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    translation TEXT NOT NULL DEFAULT '',
    position INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE entry_tags (
    entry_id TEXT NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    PRIMARY KEY (entry_id, tag)
);

CREATE TABLE sync_state (
    path TEXT PRIMARY KEY,
    last_modified DATETIME NOT NULL,
    synced_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_entries_deck ON entries(deck_id);
CREATE INDEX idx_translations_entry ON translations(entry_id);
CREATE INDEX idx_examples_entry ON examples(entry_id);

CREATE TABLE tags (
    tag TEXT PRIMARY KEY
);

CREATE TABLE deck_tags (
    deck_id TEXT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    tag TEXT NOT NULL REFERENCES tags(tag) ON DELETE CASCADE,
    PRIMARY KEY (deck_id, tag)
);

CREATE INDEX idx_deck_tags_tag ON deck_tags(tag);
