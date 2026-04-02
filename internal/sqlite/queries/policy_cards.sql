-- name: GetPolicyCard :one
-- Best-effort reader for the legacy policy card surface over the current
-- compiled policy schema.
SELECT
  COALESCE(lp.id, '') AS policy_id,
  COALESCE(s.name, '') AS service_name,
  COALESCE(le.name, '') AS log_event_name,
  CAST('' AS TEXT) AS category,
  CAST('' AS TEXT) AS category_type,
  CAST('' AS TEXT) AS "action",
  COALESCE(lp.kind, '') AS status,
  les.current_events_per_hour AS volume_per_hour,
  les.current_bytes_per_hour AS bytes_per_hour,
  les.preview_saved_total_usd_per_hour AS estimated_cost_per_hour,
  les.preview_saved_events_per_hour AS estimated_volume_per_hour,
  les.preview_saved_bytes_per_hour AS estimated_bytes_per_hour,
  COALESCE(le.severity, '') AS severity,
  CAST(NULL AS REAL) AS survival_rate,
  CAST('' AS TEXT) AS category_display_name,
  COALESCE(lp.spec, '') AS analysis,
  CAST('' AS TEXT) AS examples,
  le.baseline_avg_bytes AS event_baseline_avg_bytes,
  le.baseline_volume_per_hour AS event_baseline_volume_per_hour,
  COALESCE(lp.log_event_id, '') AS log_event_id
FROM log_event_policies lp
LEFT JOIN log_events le ON le.id = lp.log_event_id
LEFT JOIN services s ON s.id = le.service_id
LEFT JOIN log_event_statuses_cache les ON les.log_event_id = lp.log_event_id
WHERE lp.id = ?1;

-- name: ListFieldsByLogEvent :many
SELECT
  CAST('' AS TEXT) AS field_path,
  CAST(NULL AS REAL) AS baseline_avg_bytes
FROM log_event_facts
WHERE 1 = 0;
