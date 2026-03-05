-- name: GetPolicyCard :one
-- Fetches a single policy with all context needed for rich card rendering.
-- Main table: log_event_policy_statuses_cache (denormalized policy + metrics).
-- JOINs enrich with: category display name, AI analysis, log examples, baselines.
SELECT
  COALESCE(ps.policy_id, '') AS policy_id,
  COALESCE(ps.service_name, '') AS service_name,
  COALESCE(ps.log_event_name, '') AS log_event_name,
  COALESCE(ps.category, '') AS category,
  COALESCE(ps.category_type, '') AS category_type,
  COALESCE(ps.action, '') AS "action",
  COALESCE(ps.status, '') AS status,
  ps.volume_per_hour,
  ps.bytes_per_hour,
  ps.estimated_cost_reduction_per_hour_usd AS estimated_cost_per_hour,
  ps.estimated_volume_reduction_per_hour AS estimated_volume_per_hour,
  ps.estimated_bytes_reduction_per_hour AS estimated_bytes_per_hour,
  COALESCE(ps.severity, '') AS severity,
  ps.survival_rate,
  COALESCE(cat.display_name, '') AS category_display_name,
  COALESCE(lep.analysis, '') AS analysis,
  COALESCE(le.examples, '') AS examples,
  le.baseline_avg_bytes AS event_baseline_avg_bytes,
  le.baseline_volume_per_hour AS event_baseline_volume_per_hour,
  COALESCE(ps.log_event_id, '') AS log_event_id
FROM log_event_policy_statuses_cache ps
LEFT JOIN log_event_policy_category_statuses_cache cat ON cat.category = ps.category
LEFT JOIN log_event_policies lep ON lep.id = ps.policy_id
LEFT JOIN log_events le ON le.id = ps.log_event_id
WHERE ps.policy_id = ?1;

-- name: ListFieldsByLogEvent :many
-- Returns per-field metadata for a log event, used to show per-field byte impact
-- in quality policies (instrumentation_bloat, oversized_fields, duplicate_fields).
SELECT
  COALESCE(field_path, '') AS field_path,
  baseline_avg_bytes
FROM log_event_fields
WHERE log_event_id = ?1
ORDER BY baseline_avg_bytes DESC;
