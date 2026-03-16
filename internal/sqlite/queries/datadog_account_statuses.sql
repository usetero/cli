-- name: GetAccountSummary :one
SELECT
  -- ready
  CAST(COALESCE(MAX(ready_for_use), 0) AS INTEGER) AS ready_for_use,

  -- health
  COALESCE(MAX(health), '') AS health,

  -- services
  CAST(COALESCE(SUM(log_service_count), 0) AS INTEGER) AS service_count,
  CAST(COALESCE(SUM(log_active_services), 0) AS INTEGER) AS active_services,
  CAST(COALESCE(SUM(ok_services), 0) AS INTEGER) AS ok_services,
  CAST(COALESCE(SUM(disabled_services), 0) AS INTEGER) AS disabled_services,
  CAST(COALESCE(SUM(inactive_services), 0) AS INTEGER) AS inactive_services,

  -- events
  CAST(COALESCE(SUM(log_event_count), 0) AS INTEGER) AS event_count,
  CAST(COALESCE(SUM(log_event_analyzed_count), 0) AS INTEGER) AS analyzed_count,

  -- policies
  CAST(COALESCE(SUM(pending_recommendation_count), 0) AS INTEGER) AS pending_policy_count,
  CAST(COALESCE(SUM(approved_recommendation_count), 0) AS INTEGER) AS approved_policy_count,
  CAST(COALESCE(SUM(dismissed_recommendation_count), 0) AS INTEGER) AS dismissed_policy_count,
  CAST(COALESCE(SUM(policy_pending_critical_count), 0) AS INTEGER) AS policy_pending_critical_count,
  CAST(COALESCE(SUM(policy_pending_high_count), 0) AS INTEGER) AS policy_pending_high_count,
  CAST(COALESCE(SUM(policy_pending_medium_count), 0) AS INTEGER) AS policy_pending_medium_count,
  CAST(COALESCE(SUM(policy_pending_low_count), 0) AS INTEGER) AS policy_pending_low_count,

  -- estimated savings
  SUM(impact_total_usd_per_hour) AS estimated_cost_per_hour,
  SUM(impact_bytes_usd_per_hour) AS estimated_cost_per_hour_bytes,
  SUM(impact_volume_usd_per_hour) AS estimated_cost_per_hour_volume,
  SUM(impact_events_per_hour) AS estimated_volume_per_hour,
  SUM(impact_bytes_per_hour) AS estimated_bytes_per_hour,

  -- totals
  SUM(current_total_usd_per_hour) AS total_cost_per_hour,
  SUM(current_bytes_usd_per_hour) AS total_cost_per_hour_bytes,
  SUM(current_volume_usd_per_hour) AS total_cost_per_hour_volume,
  SUM(current_events_per_hour) AS total_volume_per_hour,
  SUM(current_bytes_per_hour) AS total_bytes_per_hour,

  -- service-level throughput
  SUM(current_service_events_per_hour) AS total_service_volume_per_hour,
  SUM(current_service_volume_usd_per_hour) AS total_service_cost_per_hour
FROM datadog_account_statuses_cache;
