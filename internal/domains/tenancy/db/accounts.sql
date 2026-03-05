-- name: List :many
SELECT id, name, created_at
FROM accounts
ORDER BY created_at DESC;

-- name: Create :exec
INSERT INTO accounts (id, name, created_at)
VALUES (?, ?, ?);

-- name: Delete :exec
DELETE FROM accounts
WHERE id = ?;
