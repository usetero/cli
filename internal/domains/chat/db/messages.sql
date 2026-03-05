-- name: ListByConversation :many
SELECT id, conversation_id, role, content, created_at
FROM messages
WHERE conversation_id = ?
ORDER BY created_at ASC;

-- name: Create :exec
INSERT INTO messages (id, conversation_id, role, content, created_at)
VALUES (?, ?, ?, ?, ?);

-- name: Delete :exec
DELETE FROM messages
WHERE id = ?;
