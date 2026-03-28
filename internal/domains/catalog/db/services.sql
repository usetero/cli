-- name: ListServices :many
SELECT id, account_id, name, enabled, initial_weekly_log_count, created_at
FROM services
ORDER BY name ASC;

-- name: ListServiceFacts :many
SELECT id, account_id, service_id, fact_group, fact_type, namespace, value, version, created_at
FROM service_facts
WHERE service_id = ?
ORDER BY namespace ASC, fact_type ASC, created_at DESC;

-- name: ListAllServiceFacts :many
SELECT id, account_id, service_id, fact_group, fact_type, namespace, value, version, created_at
FROM service_facts
ORDER BY service_id ASC, namespace ASC, fact_type ASC, created_at DESC;
