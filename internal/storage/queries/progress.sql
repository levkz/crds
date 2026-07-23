-- name: GetProgress :one
SELECT deck_id, entry_id, ease, interval, due, correct, incorrect
FROM progress
WHERE deck_id = ? AND entry_id = ?;

-- name: UpsertProgress :exec
INSERT INTO progress (deck_id, entry_id, ease, interval, due, correct, incorrect)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (deck_id, entry_id) DO UPDATE SET
    ease = excluded.ease,
    interval = excluded.interval,
    due = excluded.due,
    correct = excluded.correct,
    incorrect = excluded.incorrect;

-- name: GetDueCards :many
SELECT deck_id, entry_id, ease, interval, due, correct, incorrect
FROM progress
WHERE deck_id = ? AND due <= CURRENT_TIMESTAMP
ORDER BY due ASC;
