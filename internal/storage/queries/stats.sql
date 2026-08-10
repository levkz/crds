-- grade >= 2 counts a review as correct (ui.GradeGood or better).
-- name: GetDailyStats :many
SELECT
    CAST(strftime('%Y-%m-%d', reviewed_at) AS TEXT) as day,
    CAST(COUNT(*) AS INTEGER) as total_reviews,
    CAST(IFNULL(SUM(CASE WHEN grade >= 2 THEN 1 ELSE 0 END), 0) AS INTEGER) as correct_reviews
FROM reviews
GROUP BY day
ORDER BY day ASC;

-- name: GetReviewDays :many
SELECT DISTINCT CAST(strftime('%Y-%m-%d', reviewed_at) AS TEXT) as day
FROM reviews
ORDER BY day ASC;

-- name: GetTodayStatsByDecks :one
SELECT
    CAST(COUNT(*) AS INTEGER) as total_reviews,
    CAST(IFNULL(SUM(CASE WHEN grade >= 2 THEN 1 ELSE 0 END), 0) AS INTEGER) as correct_reviews
FROM reviews
WHERE reviewed_at >= datetime('now', 'start of day')
  AND deck_id IN (sqlc.slice('deck_ids'));

-- name: GetDailyStatsByDecks :many
SELECT
    CAST(strftime('%Y-%m-%d', reviewed_at) AS TEXT) as day,
    CAST(COUNT(*) AS INTEGER) as total_reviews,
    CAST(IFNULL(SUM(CASE WHEN grade >= 2 THEN 1 ELSE 0 END), 0) AS INTEGER) as correct_reviews
FROM reviews
WHERE deck_id IN (sqlc.slice('deck_ids'))
GROUP BY day
ORDER BY day ASC;

-- name: GetReviewDaysByDecks :many
SELECT DISTINCT CAST(strftime('%Y-%m-%d', reviewed_at) AS TEXT) as day
FROM reviews
WHERE deck_id IN (sqlc.slice('deck_ids'))
ORDER BY day ASC;

-- name: GetEntryStats :one
SELECT
    CAST(COUNT(*) AS INTEGER) as total_reviews,
    CAST(IFNULL(SUM(CASE WHEN strftime('%Y-%m-%d', reviewed_at) = strftime('%Y-%m-%d', 'now') THEN 1 ELSE 0 END), 0) AS INTEGER) as reviewed_today,
    CAST(IFNULL(SUM(CASE WHEN grade >= 2 THEN 1 ELSE 0 END), 0) AS INTEGER) as correct_reviews,
    CAST(IFNULL(SUM(CASE WHEN grade < 2 THEN 1 ELSE 0 END), 0) AS INTEGER) as incorrect_reviews,
    CAST(MAX(reviewed_at) AS TEXT) as last_reviewed
FROM reviews
WHERE entry_id = ?;

-- name: GetEntryDailyStats :many
SELECT
    CAST(strftime('%Y-%m-%d', reviewed_at) AS TEXT) as day,
    CAST(COUNT(*) AS INTEGER) as total_reviews,
    CAST(IFNULL(SUM(CASE WHEN grade >= 2 THEN 1 ELSE 0 END), 0) AS INTEGER) as correct_reviews
FROM reviews
WHERE entry_id = ?
GROUP BY day
ORDER BY day ASC;

-- name: GetEntryReviewDays :many
SELECT DISTINCT CAST(strftime('%Y-%m-%d', reviewed_at) AS TEXT) as day
FROM reviews
WHERE entry_id = ?
ORDER BY day ASC;
