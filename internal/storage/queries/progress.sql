-- name: GetProgress :one
SELECT deck_id, entry_id, reverse, ease, interval, due, correct, incorrect
FROM progress
WHERE deck_id = ? AND entry_id = ? AND reverse = ?;

-- name: UpsertProgress :exec
INSERT INTO progress (deck_id, entry_id, reverse, ease, interval, due, correct, incorrect)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (deck_id, entry_id, reverse) DO UPDATE SET
    ease = excluded.ease,
    interval = excluded.interval,
    due = excluded.due,
    correct = excluded.correct,
    incorrect = excluded.incorrect;

-- name: GetAllProgress :many
SELECT deck_id, entry_id, reverse, ease, interval, due, correct, incorrect
FROM progress;

-- name: GetDeckProgress :many
SELECT entry_id,
    CAST(SUM(correct) AS INTEGER) as correct,
    CAST(SUM(incorrect) AS INTEGER) as incorrect
FROM progress
WHERE deck_id = ?
GROUP BY entry_id;

-- name: GetDueCards :many
SELECT deck_id, entry_id, reverse, ease, interval, due, correct, incorrect
FROM progress
WHERE deck_id = ? AND due <= ?
ORDER BY due ASC;

-- name: ListNewEntriesByDeck :many
SELECT e.id
FROM entries e
LEFT JOIN progress p ON p.deck_id = e.deck_id AND p.entry_id = e.id
WHERE e.deck_id = ? AND p.entry_id IS NULL
ORDER BY e.position ASC;
