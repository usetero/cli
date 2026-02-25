-- name: ListCategoryStatusesByCostAndType :many
-- Pre-computed per-category rollup filtered by category_type (waste or compliance).
SELECT
  COALESCE(category, '') AS category,
  COALESCE(category_type, '') AS category_type,
  COALESCE("action", '') AS policy_action,
  COALESCE(display_name, '') AS display_name,
  COALESCE(principle, '') AS principle,
  CAST(COALESCE(pending_count, 0) AS INTEGER) AS pending_count,
  CAST(COALESCE(approved_count, 0) AS INTEGER) AS approved_count,
  CAST(COALESCE(dismissed_count, 0) AS INTEGER) AS dismissed_count,
  estimated_volume_reduction_per_hour,
  estimated_bytes_reduction_per_hour,
  estimated_cost_reduction_per_hour_usd,
  estimated_cost_reduction_per_hour_bytes_usd,
  estimated_cost_reduction_per_hour_volume_usd,
  CAST(COALESCE(events_with_volumes, 0) AS INTEGER) AS events_with_volumes,
  CAST(COALESCE(total_event_count, 0) AS INTEGER) AS total_event_count,
  CAST(COALESCE(policy_pending_critical_count, 0) AS INTEGER) AS policy_pending_critical_count,
  CAST(COALESCE(policy_pending_high_count, 0) AS INTEGER) AS policy_pending_high_count,
  CAST(COALESCE(policy_pending_medium_count, 0) AS INTEGER) AS policy_pending_medium_count,
  CAST(COALESCE(policy_pending_low_count, 0) AS INTEGER) AS policy_pending_low_count
FROM log_event_policy_category_statuses_cache
WHERE category IS NOT NULL AND category != ''
  AND category_type = ?1
ORDER BY estimated_cost_reduction_per_hour_usd DESC NULLS LAST, pending_count DESC;
