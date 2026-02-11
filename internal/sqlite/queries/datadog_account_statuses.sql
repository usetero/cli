-- name: GetCatalogStatus :one
SELECT
  CAST(COALESCE(MAX(ready_for_use), 0) AS INTEGER) AS ready_for_use,
  CAST(COALESCE(SUM(log_service_count), 0) AS INTEGER) AS service_count,
  CAST(COALESCE(SUM(log_active_services), 0) AS INTEGER) AS active_services,
  CAST(COALESCE(SUM(log_event_count), 0) AS INTEGER) AS event_count,
  CAST(COALESCE(SUM(log_analyzed_count), 0) AS INTEGER) AS analyzed_count,
  CAST(COALESCE(SUM(log_analyzing_count), 0) AS INTEGER) AS analyzing_count,
  CAST(COALESCE(SUM(log_discovering_count), 0) AS INTEGER) AS discovering_count,
  CAST(COALESCE(SUM(log_broken_services), 0) AS INTEGER) AS broken_services,
  CAST(COALESCE(SUM(log_stale_services), 0) AS INTEGER) AS stale_services,
  CAST(COALESCE(AVG(log_percent_complete), 0.0) AS REAL) AS percent_complete,
  CASE MIN(CASE log_status
    WHEN 'DISABLED' THEN 7
    WHEN 'INACTIVE' THEN 6
    WHEN 'BROKEN' THEN 1
    WHEN 'STALE' THEN 2
    WHEN 'DISCOVERING' THEN 3
    WHEN 'ANALYZING' THEN 4
    WHEN 'READY' THEN 5
    ELSE 8
  END)
    WHEN 1 THEN 'BROKEN'
    WHEN 2 THEN 'STALE'
    WHEN 3 THEN 'DISCOVERING'
    WHEN 4 THEN 'ANALYZING'
    WHEN 5 THEN 'READY'
    WHEN 6 THEN 'INACTIVE'
    WHEN 7 THEN 'DISABLED'
    ELSE ''
  END AS worst_status,
  COALESCE(MAX(log_error), '') AS log_error
FROM datadog_account_statuses_cache;

-- name: GetPolicyStatus :one
SELECT
  CAST(COALESCE(MAX(ready_for_use), 0) AS INTEGER) AS ready_for_use,
  CAST(COALESCE(SUM(log_pending_policy_count), 0) AS INTEGER) AS pending_policy_count,
  CAST(COALESCE(SUM(log_policy_count), 0) AS INTEGER) AS policy_count,
  CAST(COALESCE(SUM(log_approved_policy_count), 0) AS INTEGER) AS approved_policy_count,
  SUM(log_estimated_cost_reduction_per_hour_usd) AS estimated_cost_per_hour,
  SUM(log_estimated_cost_reduction_per_hour_bytes_usd) AS estimated_cost_per_hour_bytes,
  SUM(log_estimated_cost_reduction_per_hour_volume_usd) AS estimated_cost_per_hour_volume,
  CAST(COALESCE(SUM(log_estimated_volume_reduction_per_hour), 0.0) AS REAL) AS estimated_volume_per_hour,
  CAST(COALESCE(SUM(log_estimated_bytes_reduction_per_hour), 0.0) AS REAL) AS estimated_bytes_per_hour,
  SUM(log_observed_cost_per_hour_before_usd) AS observed_cost_before,
  SUM(log_observed_cost_per_hour_after_usd) AS observed_cost_after,
  CAST(COALESCE(SUM(log_observed_volume_per_hour_before), 0.0) AS REAL) AS observed_volume_before,
  CAST(COALESCE(SUM(log_observed_volume_per_hour_after), 0.0) AS REAL) AS observed_volume_after,
  SUM(log_cost_per_hour_usd) AS total_cost_per_hour,
  CAST(COALESCE(SUM(log_volume_per_hour), 0.0) AS REAL) AS total_volume_per_hour,
  CAST(COALESCE(SUM(log_bytes_per_hour), 0.0) AS REAL) AS total_bytes_per_hour
FROM datadog_account_statuses_cache;
