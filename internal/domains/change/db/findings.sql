-- name: ListFindings :many
SELECT
    f.id,
    f.account_id,
    f.service_id,
    f.log_event_id,
    f.domain,
    f.scope_kind,
    f.type,
    f.problem_version,
    f.fingerprint,
    f.details,
    f.closed_at,
    f.created_at,
    fc.disposition,
    fc.title,
    fc.body,
    fc.priority,
    fs.plan_status,
    fp.title AS plan_title,
    fp.summary AS plan_summary,
    fp.rationale AS plan_rationale,
    fp.open_questions AS plan_open_questions,
    fp.steps AS plan_steps,
    fp.revision AS plan_revision,
    fp.version AS plan_version,
    fs.isolated_events_per_hour,
    fs.isolated_total_usd_per_hour,
    fs.log_event_count,
    fs.finding_updated_at,
    fs.plan_updated_at
FROM findings f
LEFT JOIN finding_curations fc ON fc.id = (
    SELECT fc2.id
    FROM finding_curations fc2
    WHERE fc2.finding_id = f.id
    ORDER BY fc2.version DESC, fc2.created_at DESC
    LIMIT 1
)
LEFT JOIN finding_statuses_cache fs ON fs.finding_id = f.id
LEFT JOIN finding_plans fp ON fp.id = (
    SELECT fp2.id
    FROM finding_plans fp2
    WHERE fp2.finding_id = f.id
    ORDER BY fp2.version DESC, fp2.revision DESC, fp2.created_at DESC
    LIMIT 1
)
ORDER BY f.domain ASC, f.type ASC, f.created_at DESC;

-- name: ListFindingsByDomain :many
SELECT
    f.id,
    f.account_id,
    f.service_id,
    f.log_event_id,
    f.domain,
    f.scope_kind,
    f.type,
    f.problem_version,
    f.fingerprint,
    f.details,
    f.closed_at,
    f.created_at,
    fc.disposition,
    fc.title,
    fc.body,
    fc.priority,
    fs.plan_status,
    fp.title AS plan_title,
    fp.summary AS plan_summary,
    fp.rationale AS plan_rationale,
    fp.open_questions AS plan_open_questions,
    fp.steps AS plan_steps,
    fp.revision AS plan_revision,
    fp.version AS plan_version,
    fs.isolated_events_per_hour,
    fs.isolated_total_usd_per_hour,
    fs.log_event_count,
    fs.finding_updated_at,
    fs.plan_updated_at
FROM findings f
LEFT JOIN finding_curations fc ON fc.id = (
    SELECT fc2.id
    FROM finding_curations fc2
    WHERE fc2.finding_id = f.id
    ORDER BY fc2.version DESC, fc2.created_at DESC
    LIMIT 1
)
LEFT JOIN finding_statuses_cache fs ON fs.finding_id = f.id
LEFT JOIN finding_plans fp ON fp.id = (
    SELECT fp2.id
    FROM finding_plans fp2
    WHERE fp2.finding_id = f.id
    ORDER BY fp2.version DESC, fp2.revision DESC, fp2.created_at DESC
    LIMIT 1
)
WHERE f.domain = ?
ORDER BY f.type ASC, f.created_at DESC;

-- name: ListFindingsByService :many
SELECT
    f.id,
    f.account_id,
    f.service_id,
    f.log_event_id,
    f.domain,
    f.scope_kind,
    f.type,
    f.problem_version,
    f.fingerprint,
    f.details,
    f.closed_at,
    f.created_at,
    fc.disposition,
    fc.title,
    fc.body,
    fc.priority,
    fs.plan_status,
    fp.title AS plan_title,
    fp.summary AS plan_summary,
    fp.rationale AS plan_rationale,
    fp.open_questions AS plan_open_questions,
    fp.steps AS plan_steps,
    fp.revision AS plan_revision,
    fp.version AS plan_version,
    fs.isolated_events_per_hour,
    fs.isolated_total_usd_per_hour,
    fs.log_event_count,
    fs.finding_updated_at,
    fs.plan_updated_at
FROM findings f
LEFT JOIN finding_curations fc ON fc.id = (
    SELECT fc2.id
    FROM finding_curations fc2
    WHERE fc2.finding_id = f.id
    ORDER BY fc2.version DESC, fc2.created_at DESC
    LIMIT 1
)
LEFT JOIN finding_statuses_cache fs ON fs.finding_id = f.id
LEFT JOIN finding_plans fp ON fp.id = (
    SELECT fp2.id
    FROM finding_plans fp2
    WHERE fp2.finding_id = f.id
    ORDER BY fp2.version DESC, fp2.revision DESC, fp2.created_at DESC
    LIMIT 1
)
WHERE f.service_id = ?
ORDER BY f.domain ASC, f.type ASC, f.created_at DESC;
