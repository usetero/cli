-- name: ListTopPendingPoliciesByCategory :many
-- Legacy policy-category surface. The current backend is findings-first, so
-- this reader intentionally returns no rows until the UI is migrated.
SELECT
  CAST('' AS TEXT) AS service_name,
  CAST('' AS TEXT) AS log_event_name,
  CAST(NULL AS REAL) AS volume_per_hour,
  CAST(NULL AS REAL) AS bytes_per_hour,
  CAST(NULL AS REAL) AS estimated_cost_per_hour,
  CAST(NULL AS REAL) AS estimated_cost_per_hour_bytes,
  CAST(NULL AS REAL) AS estimated_cost_per_hour_volume,
  CAST(NULL AS REAL) AS estimated_bytes_per_hour,
  CAST(NULL AS REAL) AS estimated_volume_per_hour
FROM findings
WHERE 1 = 0;

-- name: CountFixedPIIPolicies :one
SELECT CAST(0 AS INTEGER);
