-- name: GetLastModified :one
SELECT last_modified FROM sync_state WHERE path = ?;

-- name: UpsertSyncState :exec
INSERT INTO sync_state (path, last_modified, synced_at)
VALUES (?, ?, CURRENT_TIMESTAMP)
ON CONFLICT (path) DO UPDATE SET
    last_modified = EXCLUDED.last_modified,
    synced_at = CURRENT_TIMESTAMP;
