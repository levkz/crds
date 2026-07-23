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
