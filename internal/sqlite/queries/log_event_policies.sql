-- name: CountLogEventPolicies :one
SELECT COUNT(*) FROM log_event_policies;

-- name: ListPolicyCategoryStatuses :many
SELECT
  COALESCE(category, '') AS category,
  CAST(COALESCE(SUM(CASE WHEN status = 'PENDING' THEN 1 ELSE 0 END), 0) AS INTEGER) AS pending_count,
  CAST(COALESCE(SUM(CASE WHEN status = 'APPROVED' THEN 1 ELSE 0 END), 0) AS INTEGER) AS approved_count,
  CAST(COALESCE(SUM(CASE WHEN status = 'DISMISSED' THEN 1 ELSE 0 END), 0) AS INTEGER) AS dismissed_count,
  SUM(CASE WHEN status = 'PENDING' THEN estimated_volume_reduction_per_hour ELSE 0 END) AS estimated_volume_per_hour,
  SUM(CASE WHEN status = 'PENDING' THEN estimated_bytes_reduction_per_hour ELSE 0 END) AS estimated_bytes_per_hour,
  SUM(CASE WHEN status = 'PENDING' THEN estimated_cost_reduction_per_hour_usd ELSE 0 END) AS estimated_cost_per_hour,
  CASE MAX(CASE risk_level
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
  CAST(COALESCE(GROUP_CONCAT(DISTINCT benefits), '') AS TEXT) AS benefits
FROM log_event_policy_statuses_cache
WHERE category IS NOT NULL AND category != ''
GROUP BY category
ORDER BY
  SUM(CASE WHEN status = 'PENDING' THEN 1 ELSE 0 END) DESC;
