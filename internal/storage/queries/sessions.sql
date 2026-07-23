-- name: CreateSession :one
INSERT INTO sessions (started_at)
VALUES (CURRENT_TIMESTAMP)
RETURNING id, started_at, finished_at, reviewed, correct, incorrect, duration_ms;

-- name: FinishSession :exec
UPDATE sessions
SET finished_at = CURRENT_TIMESTAMP,
    reviewed = ?,
    correct = ?,
    incorrect = ?,
    duration_ms = ?
WHERE id = ?;

-- name: GetSession :one
SELECT id, started_at, finished_at, reviewed, correct, incorrect, duration_ms
FROM sessions
WHERE id = ?;
