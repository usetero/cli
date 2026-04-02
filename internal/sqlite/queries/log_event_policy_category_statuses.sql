-- name: ListCategoryStatusesByCostAndType :many
-- Legacy category rollup surface. Findings now drive issue discovery, so keep
-- the old category tabs empty until they are reworked on top of findings.
SELECT
  CAST('' AS TEXT) AS category,
  CAST('' AS TEXT) AS category_type,
  CAST('' AS TEXT) AS policy_action,
  CAST('' AS TEXT) AS display_name,
  CAST('' AS TEXT) AS principle,
  CAST(0 AS INTEGER) AS pending_count,
  CAST(0 AS INTEGER) AS approved_count,
  CAST(0 AS INTEGER) AS dismissed_count,
  CAST(NULL AS REAL) AS estimated_volume_reduction_per_hour,
  CAST(NULL AS REAL) AS estimated_bytes_reduction_per_hour,
  CAST(NULL AS REAL) AS estimated_cost_reduction_per_hour_usd,
  CAST(NULL AS REAL) AS estimated_cost_reduction_per_hour_bytes_usd,
  CAST(NULL AS REAL) AS estimated_cost_reduction_per_hour_volume_usd,
  CAST(0 AS INTEGER) AS events_with_volumes,
  CAST(0 AS INTEGER) AS total_event_count,
  CAST(0 AS INTEGER) AS policy_pending_critical_count,
  CAST(0 AS INTEGER) AS policy_pending_high_count,
  CAST(0 AS INTEGER) AS policy_pending_medium_count,
  CAST(0 AS INTEGER) AS policy_pending_low_count
FROM findings
WHERE 1 = 0;
