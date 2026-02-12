-- name: GetService :one
SELECT * FROM services WHERE id = ?;

-- name: ListServices :many
SELECT * FROM services ORDER BY name;

-- name: ListServicesByAccount :many
SELECT * FROM services WHERE account_id = ? ORDER BY name;

-- name: CountServices :one
SELECT COUNT(*) FROM services;

-- name: SetServiceEnabled :exec
UPDATE services SET enabled = ? WHERE id = ?;
