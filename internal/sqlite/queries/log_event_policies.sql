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

-- name: ListPendingPIIPolicies :many
SELECT
  COALESCE(s.name, '') AS service_name,
  COALESCE(le.name, '') AS log_event_name,
  COALESCE(lep.analysis, '') AS analysis,
  COALESCE(les.volume_per_hour, 0) AS volume_per_hour,
  CAST(COALESCE((
    SELECT MAX(CASE json_extract(f.value, '$.pii_type')
      WHEN 'credit_card' THEN 2 WHEN 'ssn' THEN 2 WHEN 'password' THEN 2
      WHEN 'email' THEN 1 WHEN 'name' THEN 1 WHEN 'phone' THEN 1
      WHEN 'address' THEN 1 WHEN 'ip_address' THEN 1 WHEN 'date_of_birth' THEN 1
      ELSE 0 END)
    FROM json_each(json_extract(lep.analysis, '$.pii_leakage.fields')) f
  ), 0) AS INTEGER) AS max_severity
FROM log_event_policy_statuses_cache leps
JOIN log_events le ON le.id = leps.log_event_id
JOIN services s ON s.id = le.service_id
LEFT JOIN log_event_policies lep ON lep.id = leps.policy_id
LEFT JOIN log_event_statuses_cache les ON les.log_event_id = leps.log_event_id
WHERE leps.category = 'pii_leakage' AND leps.status = 'PENDING'
ORDER BY max_severity DESC, les.volume_per_hour DESC;

-- name: ListTopPendingPoliciesByCategory :many
SELECT
  COALESCE(s.name, '') AS service_name,
  COALESCE(le.name, '') AS log_event_name,
  COALESCE(les.volume_per_hour, 0) AS volume_per_hour,
  COALESCE(leps.estimated_cost_reduction_per_hour_usd, 0) AS estimated_cost_per_hour,
  COALESCE(leps.estimated_bytes_reduction_per_hour, 0) AS estimated_bytes_per_hour
FROM log_event_policy_statuses_cache leps
JOIN log_events le ON le.id = leps.log_event_id
JOIN services s ON s.id = le.service_id
LEFT JOIN log_event_statuses_cache les ON les.log_event_id = leps.log_event_id
WHERE leps.category = ?1 AND leps.status = 'PENDING'
ORDER BY leps.estimated_cost_reduction_per_hour_usd DESC, les.volume_per_hour DESC
LIMIT ?2;

-- name: CountFixedPIIPolicies :one
SELECT CAST(COUNT(*) AS INTEGER) FROM log_event_policy_statuses_cache
WHERE category = 'pii_leakage' AND status = 'APPROVED';
