-- Compliance policy queries for PII, Secrets, PHI, and Payment Data leakage.

-- name: ListComplianceCategorySummaries :many
-- Returns summary stats for each of the 4 compliance categories.
SELECT
  category,
  CAST(SUM(CASE WHEN any_observed = 1 THEN 1 ELSE 0 END) AS INTEGER) AS leaking_count,
  CAST(SUM(CASE WHEN any_observed = 0 THEN 1 ELSE 0 END) AS INTEGER) AS at_risk_count,
  CAST(SUM(approved_count) AS INTEGER) AS fixed_count,
  COALESCE(SUM(volume_per_hour), 0.0) AS volume_per_hour,
  COUNT(DISTINCT service_name) AS service_count,
  GROUP_CONCAT(DISTINCT service_name) AS unique_services
FROM (
  SELECT
    leps.category,
    s.name AS service_name,
    les.volume_per_hour,
    CAST(COALESCE((
      SELECT MAX(CASE json_extract(f.value, '$.observed') WHEN 1 THEN 1 ELSE 0 END)
      FROM json_each(json_extract(lep.analysis, '$.' || leps.category || '.fields')) f
    ), 0) AS INTEGER) AS any_observed,
    0 AS approved_count
  FROM log_event_policy_statuses_cache leps
  JOIN log_events le ON le.id = leps.log_event_id
  JOIN services s ON s.id = le.service_id
  LEFT JOIN log_event_policies lep ON lep.id = leps.policy_id
  LEFT JOIN log_event_statuses_cache les ON les.log_event_id = leps.log_event_id
  WHERE leps.category IN ('pii_leakage', 'secrets_leakage', 'phi_leakage', 'payment_data_leakage')
    AND leps.status = 'PENDING'

  UNION ALL

  SELECT
    leps.category,
    s.name AS service_name,
    0.0 AS volume_per_hour,
    0 AS any_observed,
    1 AS approved_count
  FROM log_event_policy_statuses_cache leps
  JOIN log_events le ON le.id = leps.log_event_id
  JOIN services s ON s.id = le.service_id
  WHERE leps.category IN ('pii_leakage', 'secrets_leakage', 'phi_leakage', 'payment_data_leakage')
    AND leps.status = 'APPROVED'
)
GROUP BY category
ORDER BY leaking_count DESC, at_risk_count DESC;

-- name: ListPendingCompliancePoliciesByCategory :many
-- Returns pending compliance policies for a specific category, sorted by observed then volume.
SELECT
  COALESCE(s.name, '') AS service_name,
  COALESCE(le.name, '') AS log_event_name,
  COALESCE(lep.analysis, '') AS analysis,
  les.volume_per_hour,
  CAST(COALESCE((
    SELECT MAX(CASE json_extract(f.value, '$.observed') WHEN 1 THEN 1 ELSE 0 END)
    FROM json_each(json_extract(lep.analysis, '$.' || ?1 || '.fields')) f
  ), 0) AS INTEGER) AS any_observed
FROM log_event_policy_statuses_cache leps
JOIN log_events le ON le.id = leps.log_event_id
JOIN services s ON s.id = le.service_id
LEFT JOIN log_event_policies lep ON lep.id = leps.policy_id
LEFT JOIN log_event_statuses_cache les ON les.log_event_id = leps.log_event_id
WHERE leps.category = ?1 AND leps.status = 'PENDING'
ORDER BY any_observed DESC, les.volume_per_hour DESC
LIMIT ?2;

-- name: CountTotalCompliancePolicies :one
-- Returns the total number of pending compliance policies across all 4 categories.
SELECT CAST(COUNT(*) AS INTEGER) FROM log_event_policy_statuses_cache
WHERE category IN ('pii_leakage', 'secrets_leakage', 'phi_leakage', 'payment_data_leakage')
  AND status = 'PENDING';

-- name: CountFixedCompliancePolicies :one
-- Returns the total number of approved compliance policies across all 4 categories.
SELECT CAST(COUNT(*) AS INTEGER) FROM log_event_policy_statuses_cache
WHERE category IN ('pii_leakage', 'secrets_leakage', 'phi_leakage', 'payment_data_leakage')
  AND status = 'APPROVED';
