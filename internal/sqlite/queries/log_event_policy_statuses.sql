-- name: ListTopPendingPoliciesByCategory :many
SELECT
  COALESCE(service_name, '') AS service_name,
  COALESCE(log_event_name, '') AS log_event_name,
  current_events_per_hour AS volume_per_hour,
  current_bytes_per_hour AS bytes_per_hour,
  impact_total_usd_per_hour AS estimated_cost_per_hour,
  impact_bytes_usd_per_hour AS estimated_cost_per_hour_bytes,
  impact_volume_usd_per_hour AS estimated_cost_per_hour_volume,
  impact_bytes_per_hour AS estimated_bytes_per_hour,
  impact_events_per_hour AS estimated_volume_per_hour
FROM recommendation_statuses_cache
WHERE category = ?1 AND status = 'PENDING'
ORDER BY impact_total_usd_per_hour DESC, current_events_per_hour DESC
LIMIT ?2;

-- name: ListPendingPIIPolicies :many
SELECT
  COALESCE(service_name, '') AS service_name,
  COALESCE(log_event_name, '') AS log_event_name,
  COALESCE(ler.analysis, '') AS analysis,
  current_events_per_hour AS volume_per_hour,
  CAST(COALESCE((
    SELECT MAX(CASE json_extract(f.value, '$.observed') WHEN 1 THEN 1 ELSE 0 END)
    FROM json_each(json_extract(ler.analysis, '$.pii_leakage.fields')) f
  ), 0) AS INTEGER) AS any_observed
FROM recommendation_statuses_cache leps
LEFT JOIN log_event_recommendations ler ON ler.id = leps.recommendation_id
WHERE leps.category = 'pii_leakage' AND leps.status = 'PENDING'
ORDER BY any_observed DESC, leps.current_events_per_hour DESC;

-- name: CountFixedPIIPolicies :one
SELECT CAST(COUNT(*) AS INTEGER) FROM recommendation_statuses_cache
WHERE category = 'pii_leakage' AND status = 'APPROVED';
