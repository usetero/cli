-- name: CountLogEventPolicies :one
SELECT COUNT(*) FROM log_event_recommendations;

-- name: ApproveLogEventPolicy :exec
UPDATE log_event_recommendations
SET approved_at = ?, approved_by = ?
WHERE id = ?;

-- name: DismissLogEventPolicy :exec
UPDATE log_event_recommendations
SET dismissed_at = ?, dismissed_by = ?
WHERE id = ?;
