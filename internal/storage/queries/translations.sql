-- name: InsertTranslation :exec
INSERT INTO translations (entry_id, text, position) VALUES (?, ?, ?);

-- name: DeleteTranslationsByEntry :exec
DELETE FROM translations WHERE entry_id = ?;

-- name: GetTranslationsByEntry :many
SELECT text FROM translations WHERE entry_id = ? ORDER BY position;
