-- +goose Up
CREATE TABLE tags (
    tag TEXT PRIMARY KEY
);

CREATE TABLE deck_tags (
    deck_id TEXT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    tag TEXT NOT NULL REFERENCES tags(tag) ON DELETE CASCADE,
    PRIMARY KEY (deck_id, tag)
);

CREATE INDEX idx_deck_tags_tag ON deck_tags(tag);

-- Populate tags from existing entry_tags
INSERT OR IGNORE INTO tags (tag) SELECT DISTINCT tag FROM entry_tags;

-- Populate deck_tags from existing entries + entry_tags
INSERT OR IGNORE INTO deck_tags (deck_id, tag)
SELECT DISTINCT e.deck_id, et.tag
FROM entry_tags et
JOIN entries e ON e.id = et.entry_id;

-- +goose Down
DROP TABLE IF EXISTS deck_tags;
DROP TABLE IF EXISTS tags;
