-- name: InsertTag :exec
INSERT OR IGNORE INTO tags (tag) VALUES (?);

-- name: InsertDeckTag :exec
INSERT OR IGNORE INTO deck_tags (deck_id, tag) VALUES (?, ?);

-- name: DeleteDeckTags :exec
DELETE FROM deck_tags WHERE deck_id = ?;

-- name: ListDeckTags :many
SELECT tag FROM deck_tags WHERE deck_id = ? ORDER BY tag;

-- name: ListAllTags :many
SELECT tag FROM tags ORDER BY tag;

-- name: ListDecksByTag :many
SELECT deck_id FROM deck_tags WHERE tag = ? ORDER BY deck_id;

-- name: ListDeckTagsByDecks :many
SELECT deck_id, tag FROM deck_tags
WHERE deck_id IN (/* SLICE:deck_ids*/?)
ORDER BY deck_id, tag;
