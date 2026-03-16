-- name: ListAllServiceStatuses :many
SELECT
  s.name AS service_name,
  COALESCE(ssc.health, '') AS health,
  CAST(COALESCE(ssc.log_event_count, 0) AS INTEGER) AS log_event_count,
  CAST(COALESCE(ssc.log_event_analyzed_count, 0) AS INTEGER) AS log_event_analyzed_count,
  CAST(COALESCE(ssc.pending_recommendation_count, 0) AS INTEGER) AS policy_pending_count,
  CAST(COALESCE(ssc.approved_recommendation_count, 0) AS INTEGER) AS policy_approved_count,
  CAST(COALESCE(ssc.dismissed_recommendation_count, 0) AS INTEGER) AS policy_dismissed_count,
  CAST(COALESCE(ssc.policy_pending_critical_count, 0) AS INTEGER) AS policy_pending_critical_count,
  CAST(COALESCE(ssc.policy_pending_high_count, 0) AS INTEGER) AS policy_pending_high_count,
  CAST(COALESCE(ssc.policy_pending_medium_count, 0) AS INTEGER) AS policy_pending_medium_count,
  CAST(COALESCE(ssc.policy_pending_low_count, 0) AS INTEGER) AS policy_pending_low_count,
  ssc.current_service_events_per_hour AS service_volume_per_hour,
  ssc.current_service_debug_events_per_hour AS service_debug_volume_per_hour,
  ssc.current_service_info_events_per_hour AS service_info_volume_per_hour,
  ssc.current_service_warn_events_per_hour AS service_warn_volume_per_hour,
  ssc.current_service_error_events_per_hour AS service_error_volume_per_hour,
  ssc.current_service_other_events_per_hour AS service_other_volume_per_hour,
  ssc.current_service_volume_usd_per_hour AS service_cost_per_hour_volume_usd,
  ssc.current_events_per_hour AS log_event_volume_per_hour,
  ssc.current_bytes_per_hour AS log_event_bytes_per_hour,
  ssc.current_total_usd_per_hour AS log_event_cost_per_hour_usd,
  ssc.current_bytes_usd_per_hour AS log_event_cost_per_hour_bytes_usd,
  ssc.current_volume_usd_per_hour AS log_event_cost_per_hour_volume_usd,
  ssc.impact_events_per_hour AS estimated_volume_reduction_per_hour,
  ssc.impact_bytes_per_hour AS estimated_bytes_reduction_per_hour,
  ssc.impact_total_usd_per_hour AS estimated_cost_reduction_per_hour_usd,
  ssc.impact_bytes_usd_per_hour AS estimated_cost_reduction_per_hour_bytes_usd,
  ssc.impact_volume_usd_per_hour AS estimated_cost_reduction_per_hour_volume_usd,
  ssc.current_events_per_hour AS observed_volume_per_hour_before,
  CAST((ssc.current_events_per_hour - ssc.impact_events_per_hour) AS REAL) AS observed_volume_per_hour_after,
  ssc.current_bytes_per_hour AS observed_bytes_per_hour_before,
  CAST((ssc.current_bytes_per_hour - ssc.impact_bytes_per_hour) AS REAL) AS observed_bytes_per_hour_after,
  ssc.current_total_usd_per_hour AS observed_cost_per_hour_before_usd,
  ssc.current_bytes_usd_per_hour AS observed_cost_per_hour_before_bytes_usd,
  ssc.current_volume_usd_per_hour AS observed_cost_per_hour_before_volume_usd,
  CAST((ssc.current_total_usd_per_hour - ssc.impact_total_usd_per_hour) AS REAL) AS observed_cost_per_hour_after_usd,
  CAST((ssc.current_bytes_usd_per_hour - ssc.impact_bytes_usd_per_hour) AS REAL) AS observed_cost_per_hour_after_bytes_usd,
  CAST((ssc.current_volume_usd_per_hour - ssc.impact_volume_usd_per_hour) AS REAL) AS observed_cost_per_hour_after_volume_usd
FROM service_statuses_cache ssc
JOIN services s ON ssc.service_id = s.id
ORDER BY
  CASE ssc.health
    WHEN 'OK' THEN 1
    WHEN 'DISABLED' THEN 2
    WHEN 'INACTIVE' THEN 3
    ELSE 4
  END,
  ssc.current_total_usd_per_hour DESC,
  s.name;

-- name: ListEnabledServiceStatuses :many
SELECT
  s.name AS service_name,
  COALESCE(ssc.health, '') AS health,
  CAST(COALESCE(ssc.log_event_count, 0) AS INTEGER) AS log_event_count,
  CAST(COALESCE(ssc.log_event_analyzed_count, 0) AS INTEGER) AS log_event_analyzed_count,
  CAST(COALESCE(ssc.pending_recommendation_count, 0) AS INTEGER) AS policy_pending_count,
  CAST(COALESCE(ssc.approved_recommendation_count, 0) AS INTEGER) AS policy_approved_count,
  CAST(COALESCE(ssc.dismissed_recommendation_count, 0) AS INTEGER) AS policy_dismissed_count,
  CAST(COALESCE(ssc.policy_pending_critical_count, 0) AS INTEGER) AS policy_pending_critical_count,
  CAST(COALESCE(ssc.policy_pending_high_count, 0) AS INTEGER) AS policy_pending_high_count,
  CAST(COALESCE(ssc.policy_pending_medium_count, 0) AS INTEGER) AS policy_pending_medium_count,
  CAST(COALESCE(ssc.policy_pending_low_count, 0) AS INTEGER) AS policy_pending_low_count,
  ssc.current_service_events_per_hour AS service_volume_per_hour,
  ssc.current_service_debug_events_per_hour AS service_debug_volume_per_hour,
  ssc.current_service_info_events_per_hour AS service_info_volume_per_hour,
  ssc.current_service_warn_events_per_hour AS service_warn_volume_per_hour,
  ssc.current_service_error_events_per_hour AS service_error_volume_per_hour,
  ssc.current_service_other_events_per_hour AS service_other_volume_per_hour,
  ssc.current_service_volume_usd_per_hour AS service_cost_per_hour_volume_usd,
  ssc.current_events_per_hour AS log_event_volume_per_hour,
  ssc.current_bytes_per_hour AS log_event_bytes_per_hour,
  ssc.current_total_usd_per_hour AS log_event_cost_per_hour_usd,
  ssc.current_bytes_usd_per_hour AS log_event_cost_per_hour_bytes_usd,
  ssc.current_volume_usd_per_hour AS log_event_cost_per_hour_volume_usd,
  ssc.impact_events_per_hour AS estimated_volume_reduction_per_hour,
  ssc.impact_bytes_per_hour AS estimated_bytes_reduction_per_hour,
  ssc.impact_total_usd_per_hour AS estimated_cost_reduction_per_hour_usd,
  ssc.impact_bytes_usd_per_hour AS estimated_cost_reduction_per_hour_bytes_usd,
  ssc.impact_volume_usd_per_hour AS estimated_cost_reduction_per_hour_volume_usd,
  ssc.current_events_per_hour AS observed_volume_per_hour_before,
  CAST((ssc.current_events_per_hour - ssc.impact_events_per_hour) AS REAL) AS observed_volume_per_hour_after,
  ssc.current_bytes_per_hour AS observed_bytes_per_hour_before,
  CAST((ssc.current_bytes_per_hour - ssc.impact_bytes_per_hour) AS REAL) AS observed_bytes_per_hour_after,
  ssc.current_total_usd_per_hour AS observed_cost_per_hour_before_usd,
  ssc.current_bytes_usd_per_hour AS observed_cost_per_hour_before_bytes_usd,
  ssc.current_volume_usd_per_hour AS observed_cost_per_hour_before_volume_usd,
  CAST((ssc.current_total_usd_per_hour - ssc.impact_total_usd_per_hour) AS REAL) AS observed_cost_per_hour_after_usd,
  CAST((ssc.current_bytes_usd_per_hour - ssc.impact_bytes_usd_per_hour) AS REAL) AS observed_cost_per_hour_after_bytes_usd,
  CAST((ssc.current_volume_usd_per_hour - ssc.impact_volume_usd_per_hour) AS REAL) AS observed_cost_per_hour_after_volume_usd
FROM service_statuses_cache ssc
JOIN services s ON ssc.service_id = s.id
WHERE ssc.health NOT IN ('DISABLED', 'INACTIVE')
ORDER BY
  CASE ssc.health
    WHEN 'OK' THEN 1
    ELSE 2
  END,
  ssc.current_total_usd_per_hour DESC,
  s.name
LIMIT sqlc.arg('row_limit');
