-- name: CountLogEventPolicies :one
SELECT COUNT(*) FROM log_event_policies;

-- name: ListPolicyCategoryStatuses :many
SELECT
  COALESCE(leps.category, '') AS category,
  CAST(COALESCE(SUM(CASE WHEN leps.status = 'PENDING' THEN 1 ELSE 0 END), 0) AS INTEGER) AS pending_count,
  CAST(COALESCE(SUM(CASE WHEN leps.status = 'APPROVED' THEN 1 ELSE 0 END), 0) AS INTEGER) AS approved_count,
  CAST(COALESCE(SUM(CASE WHEN leps.status = 'DISMISSED' THEN 1 ELSE 0 END), 0) AS INTEGER) AS dismissed_count,
  SUM(CASE WHEN leps.status = 'PENDING' THEN leps.estimated_volume_reduction_per_hour ELSE 0 END) AS estimated_volume_per_hour,
  SUM(CASE WHEN leps.status = 'PENDING' THEN leps.estimated_bytes_reduction_per_hour ELSE 0 END) AS estimated_bytes_per_hour,
  SUM(CASE WHEN leps.status = 'PENDING' THEN leps.estimated_cost_reduction_per_hour_usd ELSE 0 END) AS estimated_cost_per_hour,
  CASE MAX(CASE leps.risk_level
    WHEN 'high' THEN 3
    WHEN 'medium' THEN 2
    WHEN 'low' THEN 1
    ELSE 0
  END)
    WHEN 3 THEN 'high'
    WHEN 2 THEN 'medium'
    WHEN 1 THEN 'low'
    ELSE ''
  END AS risk_level,
  CAST(COALESCE(GROUP_CONCAT(DISTINCT leps.benefits), '') AS TEXT) AS benefits,
  SUM(CASE WHEN leps.status = 'APPROVED' THEN les.observed_volume_per_hour_before ELSE 0 END) AS observed_volume_before,
  SUM(CASE WHEN leps.status = 'APPROVED' THEN les.observed_volume_per_hour_after ELSE 0 END) AS observed_volume_after,
  SUM(CASE WHEN leps.status = 'APPROVED' THEN les.observed_bytes_per_hour_before ELSE 0 END) AS observed_bytes_before,
  SUM(CASE WHEN leps.status = 'APPROVED' THEN les.observed_bytes_per_hour_after ELSE 0 END) AS observed_bytes_after,
  SUM(CASE WHEN leps.status = 'APPROVED' THEN les.observed_cost_per_hour_before_usd ELSE 0 END) AS observed_cost_before,
  SUM(CASE WHEN leps.status = 'APPROVED' THEN les.observed_cost_per_hour_after_usd ELSE 0 END) AS observed_cost_after
FROM log_event_policy_statuses_cache leps
LEFT JOIN log_event_statuses_cache les ON les.log_event_id = leps.log_event_id
WHERE leps.category IS NOT NULL AND leps.category != ''
GROUP BY leps.category
ORDER BY
  SUM(CASE WHEN leps.status = 'PENDING' THEN 1 ELSE 0 END) DESC;
