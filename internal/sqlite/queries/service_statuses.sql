-- name: ListAllServiceStatuses :many
SELECT
  s.name AS service_name,
  COALESCE(ssc.health, '') AS health,
  COALESCE(ssc.error, '') AS error,
  COALESCE(ssc.error_at, '') AS error_at,
  COALESCE(ssc.warning, '') AS warning,
  COALESCE(ssc.warning_at, '') AS warning_at,
  CAST(COALESCE(ssc.log_event_count, 0) AS INTEGER) AS log_event_count,
  CAST(COALESCE(ssc.log_event_analyzed_count, 0) AS INTEGER) AS log_event_analyzed_count,
  CAST(COALESCE(ssc.log_event_quarantined_count, 0) AS INTEGER) AS log_event_quarantined_count,
  CAST(COALESCE(ssc.policy_pending_count, 0) AS INTEGER) AS policy_pending_count,
  CAST(COALESCE(ssc.policy_approved_count, 0) AS INTEGER) AS policy_approved_count,
  CAST(COALESCE(ssc.policy_dismissed_count, 0) AS INTEGER) AS policy_dismissed_count,
  COALESCE(ssc.service_volume_per_hour, 0.0) AS service_volume_per_hour,
  COALESCE(ssc.service_cost_per_hour_volume_usd, 0.0) AS service_cost_per_hour_volume_usd,
  COALESCE(ssc.log_event_volume_per_hour, 0.0) AS log_event_volume_per_hour,
  COALESCE(ssc.log_event_bytes_per_hour, 0.0) AS log_event_bytes_per_hour,
  COALESCE(ssc.log_event_cost_per_hour_usd, 0.0) AS log_event_cost_per_hour_usd,
  COALESCE(ssc.log_event_cost_per_hour_bytes_usd, 0.0) AS log_event_cost_per_hour_bytes_usd,
  COALESCE(ssc.log_event_cost_per_hour_volume_usd, 0.0) AS log_event_cost_per_hour_volume_usd,
  COALESCE(ssc.estimated_volume_reduction_per_hour, 0.0) AS estimated_volume_reduction_per_hour,
  COALESCE(ssc.estimated_bytes_reduction_per_hour, 0.0) AS estimated_bytes_reduction_per_hour,
  COALESCE(ssc.estimated_cost_reduction_per_hour_usd, 0.0) AS estimated_cost_reduction_per_hour_usd,
  COALESCE(ssc.estimated_cost_reduction_per_hour_bytes_usd, 0.0) AS estimated_cost_reduction_per_hour_bytes_usd,
  COALESCE(ssc.estimated_cost_reduction_per_hour_volume_usd, 0.0) AS estimated_cost_reduction_per_hour_volume_usd,
  COALESCE(ssc.observed_volume_per_hour_before, 0.0) AS observed_volume_per_hour_before,
  COALESCE(ssc.observed_volume_per_hour_after, 0.0) AS observed_volume_per_hour_after,
  COALESCE(ssc.observed_bytes_per_hour_before, 0.0) AS observed_bytes_per_hour_before,
  COALESCE(ssc.observed_bytes_per_hour_after, 0.0) AS observed_bytes_per_hour_after,
  COALESCE(ssc.observed_cost_per_hour_before_usd, 0.0) AS observed_cost_per_hour_before_usd,
  COALESCE(ssc.observed_cost_per_hour_before_bytes_usd, 0.0) AS observed_cost_per_hour_before_bytes_usd,
  COALESCE(ssc.observed_cost_per_hour_before_volume_usd, 0.0) AS observed_cost_per_hour_before_volume_usd,
  COALESCE(ssc.observed_cost_per_hour_after_usd, 0.0) AS observed_cost_per_hour_after_usd,
  COALESCE(ssc.observed_cost_per_hour_after_bytes_usd, 0.0) AS observed_cost_per_hour_after_bytes_usd,
  COALESCE(ssc.observed_cost_per_hour_after_volume_usd, 0.0) AS observed_cost_per_hour_after_volume_usd
FROM service_statuses_cache ssc
JOIN services s ON ssc.service_id = s.id
ORDER BY
  CASE ssc.health
    WHEN 'ERROR' THEN 1
    WHEN 'STALE' THEN 2
    WHEN 'OK' THEN 3
    WHEN 'DISABLED' THEN 4
    WHEN 'INACTIVE' THEN 5
    ELSE 6
  END,
  ssc.log_event_cost_per_hour_usd DESC,
  s.name;

-- name: ListEnabledServiceStatuses :many
SELECT
  s.name AS service_name,
  COALESCE(ssc.health, '') AS health,
  COALESCE(ssc.error, '') AS error,
  COALESCE(ssc.error_at, '') AS error_at,
  COALESCE(ssc.warning, '') AS warning,
  COALESCE(ssc.warning_at, '') AS warning_at,
  CAST(COALESCE(ssc.log_event_count, 0) AS INTEGER) AS log_event_count,
  CAST(COALESCE(ssc.log_event_analyzed_count, 0) AS INTEGER) AS log_event_analyzed_count,
  CAST(COALESCE(ssc.log_event_quarantined_count, 0) AS INTEGER) AS log_event_quarantined_count,
  CAST(COALESCE(ssc.policy_pending_count, 0) AS INTEGER) AS policy_pending_count,
  CAST(COALESCE(ssc.policy_approved_count, 0) AS INTEGER) AS policy_approved_count,
  CAST(COALESCE(ssc.policy_dismissed_count, 0) AS INTEGER) AS policy_dismissed_count,
  COALESCE(ssc.service_volume_per_hour, 0.0) AS service_volume_per_hour,
  COALESCE(ssc.service_cost_per_hour_volume_usd, 0.0) AS service_cost_per_hour_volume_usd,
  COALESCE(ssc.log_event_volume_per_hour, 0.0) AS log_event_volume_per_hour,
  COALESCE(ssc.log_event_bytes_per_hour, 0.0) AS log_event_bytes_per_hour,
  COALESCE(ssc.log_event_cost_per_hour_usd, 0.0) AS log_event_cost_per_hour_usd,
  COALESCE(ssc.log_event_cost_per_hour_bytes_usd, 0.0) AS log_event_cost_per_hour_bytes_usd,
  COALESCE(ssc.log_event_cost_per_hour_volume_usd, 0.0) AS log_event_cost_per_hour_volume_usd,
  COALESCE(ssc.estimated_volume_reduction_per_hour, 0.0) AS estimated_volume_reduction_per_hour,
  COALESCE(ssc.estimated_bytes_reduction_per_hour, 0.0) AS estimated_bytes_reduction_per_hour,
  COALESCE(ssc.estimated_cost_reduction_per_hour_usd, 0.0) AS estimated_cost_reduction_per_hour_usd,
  COALESCE(ssc.estimated_cost_reduction_per_hour_bytes_usd, 0.0) AS estimated_cost_reduction_per_hour_bytes_usd,
  COALESCE(ssc.estimated_cost_reduction_per_hour_volume_usd, 0.0) AS estimated_cost_reduction_per_hour_volume_usd,
  COALESCE(ssc.observed_volume_per_hour_before, 0.0) AS observed_volume_per_hour_before,
  COALESCE(ssc.observed_volume_per_hour_after, 0.0) AS observed_volume_per_hour_after,
  COALESCE(ssc.observed_bytes_per_hour_before, 0.0) AS observed_bytes_per_hour_before,
  COALESCE(ssc.observed_bytes_per_hour_after, 0.0) AS observed_bytes_per_hour_after,
  COALESCE(ssc.observed_cost_per_hour_before_usd, 0.0) AS observed_cost_per_hour_before_usd,
  COALESCE(ssc.observed_cost_per_hour_before_bytes_usd, 0.0) AS observed_cost_per_hour_before_bytes_usd,
  COALESCE(ssc.observed_cost_per_hour_before_volume_usd, 0.0) AS observed_cost_per_hour_before_volume_usd,
  COALESCE(ssc.observed_cost_per_hour_after_usd, 0.0) AS observed_cost_per_hour_after_usd,
  COALESCE(ssc.observed_cost_per_hour_after_bytes_usd, 0.0) AS observed_cost_per_hour_after_bytes_usd,
  COALESCE(ssc.observed_cost_per_hour_after_volume_usd, 0.0) AS observed_cost_per_hour_after_volume_usd
FROM service_statuses_cache ssc
JOIN services s ON ssc.service_id = s.id
WHERE ssc.health NOT IN ('DISABLED', 'INACTIVE')
ORDER BY
  CASE ssc.health
    WHEN 'ERROR' THEN 1
    WHEN 'STALE' THEN 2
    WHEN 'OK' THEN 3
    ELSE 4
  END,
  ssc.log_event_cost_per_hour_usd DESC,
  s.name
LIMIT sqlc.arg('row_limit');
