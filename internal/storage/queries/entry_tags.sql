-- name: InsertEntryTag :exec
INSERT INTO entry_tags (entry_id, tag) VALUES (?, ?) ON CONFLICT DO NOTHING;

-- name: DeleteTagsByEntry :exec
DELETE FROM entry_tags WHERE entry_id = ?;

-- name: GetTagsByEntry :many
SELECT tag FROM entry_tags WHERE entry_id = ? ORDER BY tag;
