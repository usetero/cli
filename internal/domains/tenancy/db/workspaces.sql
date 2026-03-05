-- name: ListByAccount :many
SELECT id, account_id, name, created_at
FROM workspaces
WHERE account_id = ?
ORDER BY created_at DESC;

-- name: Create :exec
INSERT INTO workspaces (id, account_id, name, created_at)
VALUES (?, ?, ?, ?);

-- name: Delete :exec
DELETE FROM workspaces
WHERE id = ?;
