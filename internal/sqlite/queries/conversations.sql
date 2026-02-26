-- name: GetConversation :one
SELECT * FROM conversations WHERE id = ?;

-- name: ListConversationsByAccount :many
SELECT * FROM conversations
WHERE account_id = ?
ORDER BY created_at DESC;

-- name: GetLatestConversationByAccount :one
SELECT * FROM conversations
WHERE account_id = ?
ORDER BY created_at DESC
LIMIT 1;

-- name: CountConversations :one
SELECT COUNT(*) FROM conversations;

-- name: InsertConversation :exec
INSERT INTO conversations (id, account_id, workspace_id, created_at)
VALUES (?, ?, ?, ?);

-- name: UpdateConversationTitle :exec
UPDATE conversations SET title = ? WHERE id = ?;
