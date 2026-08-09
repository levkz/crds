-- name: UpsertEntry :exec
INSERT INTO entries (id, deck_id, term, notes, position, updated_at)
VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO UPDATE SET
    deck_id = EXCLUDED.deck_id,
    term = EXCLUDED.term,
    notes = EXCLUDED.notes,
    position = EXCLUDED.position,
    updated_at = CURRENT_TIMESTAMP;

-- name: ListEntriesByDeck :many
SELECT * FROM entries WHERE deck_id = ? ORDER BY position;

-- name: GetAllEntries :many
SELECT * FROM entries ORDER BY position;

-- name: GetEntry :one
SELECT * FROM entries WHERE id = ?;

-- name: DeleteEntriesByDeck :exec
DELETE FROM entries WHERE deck_id = ?;
