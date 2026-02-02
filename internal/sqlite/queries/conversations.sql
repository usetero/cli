-- name: GetConversation :one
SELECT * FROM conversations WHERE id = ?;

-- name: ListConversationsByAccount :many
SELECT * FROM conversations
WHERE account_id = ?
ORDER BY updated_at DESC;

-- name: GetLatestConversationByAccount :one
SELECT * FROM conversations
WHERE account_id = ?
ORDER BY updated_at DESC
LIMIT 1;

-- name: InsertConversation :exec
INSERT INTO conversations (id, account_id, workspace_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?);
