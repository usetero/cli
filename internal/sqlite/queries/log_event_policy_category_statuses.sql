-- name: ListWastePolicyCategoryStatuses :many
-- Pre-computed per-category rollup for the waste tab.
-- Replaces the old GROUP BY over log_event_policy_statuses_cache.
SELECT
  COALESCE(category, '') AS category,
  COALESCE(category_type, '') AS category_type,
  COALESCE(impact_type, '') AS impact_type,
  CAST(COALESCE(pending_count, 0) AS INTEGER) AS pending_count,
  CAST(COALESCE(approved_count, 0) AS INTEGER) AS approved_count,
  CAST(COALESCE(dismissed_count, 0) AS INTEGER) AS dismissed_count,
  estimated_volume_reduction_per_hour,
  estimated_bytes_reduction_per_hour,
  estimated_cost_reduction_per_hour_usd,
  estimated_cost_reduction_per_hour_bytes_usd,
  estimated_cost_reduction_per_hour_volume_usd,
  CAST(COALESCE(events_with_volumes, 0) AS INTEGER) AS events_with_volumes,
  CAST(COALESCE(total_event_count, 0) AS INTEGER) AS total_event_count
FROM log_event_policy_category_statuses_cache
WHERE category IS NOT NULL AND category != ''
  AND category_type = 'waste'
ORDER BY pending_count DESC;
