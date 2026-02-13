-- name: GetAccountSummary :one
SELECT
  -- ready
  CAST(COALESCE(MAX(ready_for_use), 0) AS INTEGER) AS ready_for_use,

  -- health
  COALESCE(MAX(health), '') AS health,
  COALESCE(MAX(error), '') AS error,
  COALESCE(MAX(error_at), '') AS error_at,
  COALESCE(MAX(warning), '') AS warning,
  COALESCE(MAX(warning_at), '') AS warning_at,

  -- services
  CAST(COALESCE(SUM(log_service_count), 0) AS INTEGER) AS service_count,
  CAST(COALESCE(SUM(log_active_services), 0) AS INTEGER) AS active_services,
  CAST(COALESCE(SUM(ok_services), 0) AS INTEGER) AS ok_services,
  CAST(COALESCE(SUM(error_services), 0) AS INTEGER) AS error_services,
  CAST(COALESCE(SUM(stale_services), 0) AS INTEGER) AS stale_services,
  CAST(COALESCE(SUM(disabled_services), 0) AS INTEGER) AS disabled_services,
  CAST(COALESCE(SUM(inactive_services), 0) AS INTEGER) AS inactive_services,

  -- events
  CAST(COALESCE(SUM(log_event_count), 0) AS INTEGER) AS event_count,
  CAST(COALESCE(SUM(log_event_analyzed_count), 0) AS INTEGER) AS analyzed_count,
  CAST(COALESCE(SUM(log_event_quarantined_count), 0) AS INTEGER) AS quarantined_count,

  -- policies
  CAST(COALESCE(SUM(policy_pending_count), 0) AS INTEGER) AS pending_policy_count,
  CAST(COALESCE(SUM(policy_approved_count), 0) AS INTEGER) AS approved_policy_count,
  CAST(COALESCE(SUM(policy_dismissed_count), 0) AS INTEGER) AS dismissed_policy_count,

  -- estimated savings
  SUM(estimated_cost_reduction_per_hour_usd) AS estimated_cost_per_hour,
  SUM(estimated_cost_reduction_per_hour_bytes_usd) AS estimated_cost_per_hour_bytes,
  SUM(estimated_cost_reduction_per_hour_volume_usd) AS estimated_cost_per_hour_volume,
  CAST(COALESCE(SUM(estimated_volume_reduction_per_hour), 0.0) AS REAL) AS estimated_volume_per_hour,
  CAST(COALESCE(SUM(estimated_bytes_reduction_per_hour), 0.0) AS REAL) AS estimated_bytes_per_hour,

  -- observed impact
  SUM(observed_cost_per_hour_before_usd) AS observed_cost_before,
  SUM(observed_cost_per_hour_after_usd) AS observed_cost_after,
  CAST(COALESCE(SUM(observed_volume_per_hour_before), 0.0) AS REAL) AS observed_volume_before,
  CAST(COALESCE(SUM(observed_volume_per_hour_after), 0.0) AS REAL) AS observed_volume_after,
  CAST(COALESCE(SUM(observed_bytes_per_hour_before), 0.0) AS REAL) AS observed_bytes_before,
  CAST(COALESCE(SUM(observed_bytes_per_hour_after), 0.0) AS REAL) AS observed_bytes_after,

  -- totals
  SUM(log_event_cost_per_hour_usd) AS total_cost_per_hour,
  SUM(log_event_cost_per_hour_bytes_usd) AS total_cost_per_hour_bytes,
  SUM(log_event_cost_per_hour_volume_usd) AS total_cost_per_hour_volume,
  CAST(COALESCE(SUM(log_event_volume_per_hour), 0.0) AS REAL) AS total_volume_per_hour,
  CAST(COALESCE(SUM(log_event_bytes_per_hour), 0.0) AS REAL) AS total_bytes_per_hour,

  -- service-level throughput
  SUM(service_volume_per_hour) AS total_service_volume_per_hour,
  SUM(service_cost_per_hour_volume_usd) AS total_service_cost_per_hour
FROM datadog_account_statuses_cache;
