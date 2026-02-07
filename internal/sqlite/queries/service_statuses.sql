-- name: ListServiceStatuses :many
SELECT
  s.name AS service_name,
  COALESCE(ssc.log_status, '') AS log_status,
  COALESCE(ssc.log_error, '') AS log_error,
  COALESCE(ssc.log_percent_complete, 0.0) AS log_percent_complete,
  CAST(COALESCE(ssc.log_event_count, 0) AS INTEGER) AS log_event_count,
  CAST(COALESCE(ssc.log_analyzed_count, 0) AS INTEGER) AS log_analyzed_count,
  COALESCE(ssc.log_volume_per_hour, 0.0) AS log_volume_per_hour,
  COALESCE(ssc.log_bytes_per_hour, 0.0) AS log_bytes_per_hour,
  COALESCE(ssc.log_cost_per_hour_usd, 0.0) AS log_cost_per_hour_usd
FROM service_statuses_cache ssc
JOIN services s ON ssc.service_id = s.id
ORDER BY
  CASE ssc.log_status
    WHEN 'BROKEN' THEN 1
    WHEN 'STALE' THEN 2
    WHEN 'DISCOVERING' THEN 3
    WHEN 'ANALYZING' THEN 4
    WHEN 'DISABLED' THEN 5
    WHEN 'INACTIVE' THEN 6
    WHEN 'READY' THEN 7
    ELSE 8
  END,
  s.name;
