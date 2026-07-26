-- name: UpsertDeck :exec
INSERT INTO decks (id, name, language, translation_language, updated_at)
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    language = EXCLUDED.language,
    translation_language = EXCLUDED.translation_language,
    updated_at = CURRENT_TIMESTAMP;

-- name: ListDeckNames :many
SELECT id, name FROM decks ORDER BY name;

-- name: GetDeck :one
SELECT * FROM decks WHERE id = ?;

-- name: GetDeckEntryCount :one
SELECT CAST(COUNT(*) AS INTEGER) FROM entries WHERE deck_id = ?;

-- name: ListDecksWithStats :many
SELECT d.id, d.name, d.language, d.translation_language,
       COUNT(e.id) AS entry_count
FROM decks d
LEFT JOIN entries e ON e.deck_id = d.id
GROUP BY d.id
ORDER BY d.name;
