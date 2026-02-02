-- name: GetMessage :one
SELECT * FROM messages WHERE id = ?;

-- name: ListMessagesByConversation :many
SELECT * FROM messages
WHERE conversation_id = ?
ORDER BY created_at ASC;

-- name: ListMessagesByConversationDesc :many
SELECT * FROM messages
WHERE conversation_id = ?
ORDER BY created_at DESC;

-- name: GetLatestMessageByConversation :one
SELECT * FROM messages
WHERE conversation_id = ?
ORDER BY created_at DESC
LIMIT 1;

-- name: CountMessagesByConversation :one
SELECT COUNT(*) FROM messages WHERE conversation_id = ?;

-- name: InsertMessage :exec
INSERT INTO messages (id, account_id, content, conversation_id, created_at, model, role, stop_reason)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateMessageContent :exec
UPDATE messages SET content = ? WHERE id = ?;

-- name: UpdateMessageMeta :exec
UPDATE messages SET model = ?, stop_reason = ? WHERE id = ?;
