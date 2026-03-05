-- name: ListConversations :many
SELECT id, title, created_at
FROM conversations
ORDER BY created_at DESC;
