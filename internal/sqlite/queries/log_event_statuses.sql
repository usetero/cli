-- name: ListLogEventStatusesByService :many
SELECT
  COALESCE(le.name, '') AS log_event_name,
  les.current_events_per_hour AS volume_per_hour,
  les.current_bytes_per_hour AS bytes_per_hour,
  les.current_total_usd_per_hour AS cost_per_hour_usd,
  CAST(COALESCE(les.pending_recommendation_count, 0) AS INTEGER) AS pending_policy_count,
  CAST(COALESCE(les.approved_recommendation_count, 0) AS INTEGER) AS approved_policy_count,
  CAST(COALESCE(les.policy_pending_critical_count, 0) AS INTEGER) AS policy_pending_critical_count,
  CAST(COALESCE(les.policy_pending_high_count, 0) AS INTEGER) AS policy_pending_high_count,
  CAST(COALESCE(les.policy_pending_medium_count, 0) AS INTEGER) AS policy_pending_medium_count,
  CAST(COALESCE(les.policy_pending_low_count, 0) AS INTEGER) AS policy_pending_low_count
FROM log_events le
JOIN services s ON s.id = le.service_id
LEFT JOIN log_event_statuses_cache les ON les.log_event_id = le.id
WHERE s.name = ?1
ORDER BY les.current_total_usd_per_hour DESC, les.current_events_per_hour DESC
LIMIT ?2;
