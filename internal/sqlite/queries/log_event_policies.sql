-- name: CountLogEventPolicies :one
SELECT COUNT(*) FROM log_event_policies;

-- name: ListPendingPIIPolicies :many
SELECT
  COALESCE(s.name, '') AS service_name,
  COALESCE(le.name, '') AS log_event_name,
  COALESCE(lep.analysis, '') AS analysis,
  les.volume_per_hour,
  CAST(COALESCE((
    SELECT MAX(CASE json_extract(f.value, '$.observed') WHEN 1 THEN 1 ELSE 0 END)
    FROM json_each(json_extract(lep.analysis, '$.pii_leakage.fields')) f
  ), 0) AS INTEGER) AS any_observed
FROM log_event_policy_statuses_cache leps
JOIN log_events le ON le.id = leps.log_event_id
JOIN services s ON s.id = le.service_id
LEFT JOIN log_event_policies lep ON lep.id = leps.policy_id
LEFT JOIN log_event_statuses_cache les ON les.log_event_id = leps.log_event_id
WHERE leps.category = 'pii_leakage' AND leps.status = 'PENDING'
ORDER BY any_observed DESC, les.volume_per_hour DESC;

-- name: ListTopPendingPoliciesByCategory :many
SELECT
  COALESCE(s.name, '') AS service_name,
  COALESCE(le.name, '') AS log_event_name,
  les.volume_per_hour,
  les.bytes_per_hour,
  leps.estimated_cost_reduction_per_hour_usd AS estimated_cost_per_hour,
  leps.estimated_cost_reduction_per_hour_bytes_usd AS estimated_cost_per_hour_bytes,
  leps.estimated_cost_reduction_per_hour_volume_usd AS estimated_cost_per_hour_volume,
  leps.estimated_bytes_reduction_per_hour AS estimated_bytes_per_hour,
  leps.estimated_volume_reduction_per_hour AS estimated_volume_per_hour
FROM log_event_policy_statuses_cache leps
JOIN log_events le ON le.id = leps.log_event_id
JOIN services s ON s.id = le.service_id
LEFT JOIN log_event_statuses_cache les ON les.log_event_id = leps.log_event_id
WHERE leps.category = ?1 AND leps.status = 'PENDING'
ORDER BY leps.estimated_cost_reduction_per_hour_usd DESC, les.volume_per_hour DESC
LIMIT ?2;

-- name: CountFixedPIIPolicies :one
SELECT CAST(COUNT(*) AS INTEGER) FROM log_event_policy_statuses_cache
WHERE category = 'pii_leakage' AND status = 'APPROVED';
