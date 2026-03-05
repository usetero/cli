-- name: List :many
SELECT id, title, created_at
FROM conversations
ORDER BY created_at DESC;

-- name: Create :exec
INSERT INTO conversations (id, title, created_at)
VALUES (?, ?, ?);

-- name: Delete :exec
DELETE FROM conversations
WHERE id = ?;
