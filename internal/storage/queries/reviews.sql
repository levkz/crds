-- name: CreateReview :one
INSERT INTO reviews (session_id, deck_id, entry_id, grade, reverse, reviewed_at)
VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
RETURNING id, session_id, deck_id, entry_id, grade, reverse, reviewed_at;

-- name: GetReviewsBySession :many
SELECT id, session_id, deck_id, entry_id, grade, reverse, reviewed_at
FROM reviews
WHERE session_id = ?
ORDER BY id ASC;

-- name: GetReviewsByEntry :many
SELECT id, session_id, deck_id, entry_id, grade, reverse, reviewed_at
FROM reviews
WHERE entry_id = ?
ORDER BY id DESC
LIMIT ?;

-- name: GetTodayStats :one
SELECT
    CAST(COUNT(*) AS INTEGER) as total_reviews,
    CAST(IFNULL(SUM(CASE WHEN grade >= 3 THEN 1 ELSE 0 END), 0) AS INTEGER) as correct_reviews
FROM reviews
WHERE reviewed_at >= datetime('now', 'start of day');

-- name: GetTodayStatsByDeck :one
SELECT
    CAST(COUNT(*) AS INTEGER) as total_reviews,
    CAST(IFNULL(SUM(CASE WHEN grade >= 3 THEN 1 ELSE 0 END), 0) AS INTEGER) as correct_reviews
FROM reviews
WHERE reviewed_at >= datetime('now', 'start of day')
  AND deck_id = ?;

-- name: GetWeakTypingEntries :many
SELECT
    r.entry_id,
    td.user_input,
    td.correct_answer,
    td.similarity,
    r.reviewed_at,
    r.grade
FROM reviews r
JOIN typing_details td ON r.id = td.review_id
WHERE r.deck_id = ? AND r.grade < 3
ORDER BY td.similarity ASC
LIMIT ?;
