-- Compliance policy queries for PII, Secrets, PHI, and Payment Data leakage.

-- name: ListPendingCompliancePoliciesByCategory :many
-- Returns pending compliance policies for a specific category, sorted by observed then volume.
SELECT
  COALESCE(s.name, '') AS service_name,
  COALESCE(le.name, '') AS log_event_name,
  COALESCE(lep.analysis, '') AS analysis,
  les.current_events_per_hour AS volume_per_hour,
  CAST(COALESCE((
    SELECT MAX(CASE json_extract(f.value, '$.observed') WHEN 1 THEN 1 ELSE 0 END)
    FROM json_each(json_extract(lep.analysis, '$.' || ?1 || '.fields')) f
  ), 0) AS INTEGER) AS any_observed
FROM recommendation_statuses_cache leps
JOIN log_events le ON le.id = leps.log_event_id
JOIN services s ON s.id = le.service_id
LEFT JOIN log_event_recommendations lep ON lep.id = leps.recommendation_id
LEFT JOIN log_event_statuses_cache les ON les.log_event_id = leps.log_event_id
WHERE leps.category = ?1 AND leps.status = 'PENDING'
ORDER BY any_observed DESC, les.current_events_per_hour DESC
LIMIT ?2;

-- name: CountObservedPoliciesByComplianceCategory :many
-- Returns, per compliance category, how many pending policies have observed (leaking) data.
SELECT
  leps.category,
  CAST(SUM(CASE WHEN COALESCE((
    SELECT MAX(CASE json_extract(f.value, '$.observed') WHEN 1 THEN 1 ELSE 0 END)
    FROM json_each(json_extract(lep.analysis, '$.' || leps.category || '.fields')) f
  ), 0) = 1 THEN 1 ELSE 0 END) AS INTEGER) AS observed_count
FROM recommendation_statuses_cache leps
LEFT JOIN log_event_recommendations lep ON lep.id = leps.recommendation_id
WHERE leps.category_type = 'compliance' AND leps.status = 'PENDING'
GROUP BY leps.category;
