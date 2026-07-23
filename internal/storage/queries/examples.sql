-- name: InsertExample :exec
INSERT INTO examples (entry_id, text, translation, position) VALUES (?, ?, ?, ?);

-- name: DeleteExamplesByEntry :exec
DELETE FROM examples WHERE entry_id = ?;

-- name: GetExamplesByEntry :many
SELECT text, translation FROM examples WHERE entry_id = ? ORDER BY position;
