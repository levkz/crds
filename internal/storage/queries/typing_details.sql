-- name: CreateTypingDetail :exec
INSERT INTO typing_details (review_id, user_input, correct_answer, similarity)
VALUES (?, ?, ?, ?);

-- name: GetTypingDetailByReview :one
SELECT review_id, user_input, correct_answer, similarity
FROM typing_details
WHERE review_id = ?;
