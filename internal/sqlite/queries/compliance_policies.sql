-- Compliance policy readers are retained for legacy UI surfaces only.
-- The current backend no longer materializes these category-level policy rows.

-- name: ListPendingCompliancePoliciesByCategory :many
SELECT
  CAST('' AS TEXT) AS service_name,
  CAST('' AS TEXT) AS log_event_name,
  CAST('' AS TEXT) AS analysis,
  CAST(NULL AS REAL) AS volume_per_hour,
  CAST(0 AS INTEGER) AS any_observed
FROM findings
WHERE 1 = 0;

-- name: CountObservedPoliciesByComplianceCategory :many
SELECT
  CAST(NULL AS TEXT) AS category,
  CAST(0 AS INTEGER) AS observed_count
FROM findings
WHERE 1 = 0;
