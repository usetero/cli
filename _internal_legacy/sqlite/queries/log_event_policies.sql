-- name: CountLogEventPolicies :one
SELECT COUNT(*) FROM log_event_policies;

-- name: ApproveLogEventPolicy :exec
UPDATE log_event_policies
SET approved_at = ?, approved_by = ?
WHERE id = ?;

-- name: DismissLogEventPolicy :exec
UPDATE log_event_policies
SET dismissed_at = ?, dismissed_by = ?
WHERE id = ?;
