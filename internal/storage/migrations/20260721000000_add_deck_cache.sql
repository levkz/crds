-- +goose Up

CREATE TABLE decks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    language TEXT NOT NULL,
    translation_language TEXT NOT NULL,
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

-- +goose Down
DROP INDEX IF EXISTS idx_examples_entry;
DROP INDEX IF EXISTS idx_translations_entry;
DROP INDEX IF EXISTS idx_entries_deck;
DROP TABLE IF EXISTS sync_state;
DROP TABLE IF EXISTS entry_tags;
DROP TABLE IF EXISTS examples;
DROP TABLE IF EXISTS translations;
DROP TABLE IF EXISTS entries;
DROP TABLE IF EXISTS decks;
