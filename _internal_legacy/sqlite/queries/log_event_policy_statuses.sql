-- name: ListTopPendingPoliciesByCategory :many
SELECT
  COALESCE(service_name, '') AS service_name,
  COALESCE(log_event_name, '') AS log_event_name,
  volume_per_hour,
  bytes_per_hour,
  estimated_cost_reduction_per_hour_usd AS estimated_cost_per_hour,
  estimated_cost_reduction_per_hour_bytes_usd AS estimated_cost_per_hour_bytes,
  estimated_cost_reduction_per_hour_volume_usd AS estimated_cost_per_hour_volume,
  estimated_bytes_reduction_per_hour AS estimated_bytes_per_hour,
  estimated_volume_reduction_per_hour AS estimated_volume_per_hour
FROM log_event_policy_statuses_cache
WHERE category = ?1 AND status = 'PENDING'
ORDER BY estimated_cost_reduction_per_hour_usd DESC, volume_per_hour DESC
LIMIT ?2;

-- name: ListPendingPIIPolicies :many
SELECT
  COALESCE(service_name, '') AS service_name,
  COALESCE(log_event_name, '') AS log_event_name,
  COALESCE(lep.analysis, '') AS analysis,
  volume_per_hour,
  CAST(COALESCE((
    SELECT MAX(CASE json_extract(f.value, '$.observed') WHEN 1 THEN 1 ELSE 0 END)
    FROM json_each(json_extract(lep.analysis, '$.pii_leakage.fields')) f
  ), 0) AS INTEGER) AS any_observed
FROM log_event_policy_statuses_cache leps
LEFT JOIN log_event_policies lep ON lep.id = leps.policy_id
WHERE leps.category = 'pii_leakage' AND leps.status = 'PENDING'
ORDER BY any_observed DESC, leps.volume_per_hour DESC;

-- name: CountFixedPIIPolicies :one
SELECT CAST(COUNT(*) AS INTEGER) FROM log_event_policy_statuses_cache
WHERE category = 'pii_leakage' AND status = 'APPROVED';
