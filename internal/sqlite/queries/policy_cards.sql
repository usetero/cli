-- name: GetPolicyCard :one
-- Fetches a single policy with all context needed for rich card rendering.
-- Main table: recommendation_statuses_cache (denormalized policy + metrics).
-- JOINs enrich with: category display name, AI analysis, log examples, baselines.
SELECT
  COALESCE(ps.recommendation_id, '') AS policy_id,
  COALESCE(ps.service_name, '') AS service_name,
  COALESCE(ps.log_event_name, '') AS log_event_name,
  COALESCE(ps.category, '') AS category,
  COALESCE(ps.category_type, '') AS category_type,
  COALESCE(ps.action, '') AS "action",
  COALESCE(ps.status, '') AS status,
  ps.current_events_per_hour AS volume_per_hour,
  ps.current_bytes_per_hour AS bytes_per_hour,
  ps.impact_total_usd_per_hour AS estimated_cost_per_hour,
  ps.impact_events_per_hour AS estimated_volume_per_hour,
  ps.impact_bytes_per_hour AS estimated_bytes_per_hour,
  COALESCE(ps.severity, '') AS severity,
  ps.survival_rate,
  COALESCE(cat.display_name, '') AS category_display_name,
  COALESCE(lep.analysis, '') AS analysis,
  COALESCE(le.examples, '') AS examples,
  le.baseline_avg_bytes AS event_baseline_avg_bytes,
  le.baseline_volume_per_hour AS event_baseline_volume_per_hour,
  COALESCE(ps.log_event_id, '') AS log_event_id
FROM recommendation_statuses_cache ps
LEFT JOIN recommendation_category_statuses_cache cat ON cat.category = ps.category
LEFT JOIN log_event_recommendations lep ON lep.id = ps.recommendation_id
LEFT JOIN log_events le ON le.id = ps.log_event_id
WHERE ps.recommendation_id = ?1;

-- name: ListFieldsByLogEvent :many
-- Returns per-field metadata for a log event, used to show per-field byte impact
-- in quality policies (instrumentation_bloat, oversized_fields, duplicate_fields).
SELECT
  CAST('' AS TEXT) AS field_path,
  CAST(0 AS REAL) AS baseline_avg_bytes
WHERE 1 = 0;
