-- name: ListLogEventsByService :many
SELECT id, account_id, service_id, name, description, severity, baseline_avg_bytes, baseline_volume_per_hour, created_at
FROM log_events
WHERE service_id = ?
ORDER BY name ASC;

-- name: ListLogEventFacts :many
SELECT id, account_id, log_event_id, fact_name, slice_name, slice_version, value, created_at
FROM log_event_facts
WHERE log_event_id = ?
ORDER BY slice_name ASC, fact_name ASC, created_at DESC;

-- name: ListLogEventFactsByService :many
SELECT lef.id, lef.account_id, lef.log_event_id, lef.fact_name, lef.slice_name, lef.slice_version, lef.value, lef.created_at
FROM log_event_facts lef
JOIN log_events le ON le.id = lef.log_event_id
WHERE le.service_id = ?
ORDER BY lef.log_event_id ASC, lef.slice_name ASC, lef.fact_name ASC, lef.created_at DESC;
