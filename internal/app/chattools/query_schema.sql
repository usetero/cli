-- Tero Client Schema
--
-- This describes the local SQLite database synced from the Tero control plane.
-- All data is scoped to the authenticated user's account.
--
-- Key concepts:
--   services      - Applications producing logs (e.g., 'checkout-service')
--   log_events    - Distinct event patterns within a service
--   policies      - AI-identified waste (health checks, duplicate fields, bloat)
--   *_cache       - Pre-computed status and metrics (query these for current state)
--
-- All queries are READ-ONLY. This is a local sync of server data.

-- Chat session between a user and the AI assistant within an account.
CREATE TABLE conversations (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Account this conversation belongs to
    created_at TEXT, -- When the conversation was created
    title TEXT, -- AI-generated title, set after first exchange
    user_id TEXT -- WorkOS user ID who owns this conversation
);

-- Cache table for canonical datadog_account_statuses view. Refreshed by worker-owned status loops.
CREATE TABLE datadog_account_statuses_cache (
    id TEXT,
    account_id TEXT,
    current_bytes_per_hour REAL,
    current_bytes_usd_per_hour REAL,
    current_events_per_hour REAL,
    current_service_events_per_hour REAL,
    current_service_volume_usd_per_hour REAL,
    current_total_usd_per_hour REAL,
    current_volume_usd_per_hour REAL,
    datadog_account_id TEXT,
    disabled_services INTEGER,
    effective_bytes_per_hour REAL,
    effective_bytes_usd_per_hour REAL,
    effective_events_per_hour REAL,
    effective_log_event_count INTEGER,
    effective_saved_bytes_per_hour REAL,
    effective_saved_bytes_usd_per_hour REAL,
    effective_saved_events_per_hour REAL,
    effective_saved_total_usd_per_hour REAL,
    effective_saved_volume_usd_per_hour REAL,
    effective_total_usd_per_hour REAL,
    effective_volume_usd_per_hour REAL,
    health TEXT, -- Health state for this status row. Values: DISABLED = The user turned the log pipeline off.; INACTIVE = The service currently has no log volume.; ERROR = The log pipeline is unhealthy or failing.; OK = The log pipeline is healthy.
    inactive_services INTEGER,
    log_active_services INTEGER,
    log_event_analyzed_count INTEGER,
    log_event_count INTEGER,
    log_service_count INTEGER,
    ok_services INTEGER,
    preview_bytes_per_hour REAL,
    preview_bytes_usd_per_hour REAL,
    preview_events_per_hour REAL,
    preview_log_event_count INTEGER,
    preview_saved_bytes_per_hour REAL,
    preview_saved_bytes_usd_per_hour REAL,
    preview_saved_events_per_hour REAL,
    preview_saved_total_usd_per_hour REAL,
    preview_saved_volume_usd_per_hour REAL,
    preview_total_usd_per_hour REAL,
    preview_volume_usd_per_hour REAL,
    ready_for_use INTEGER
);

-- Datadog integration configuration for an account, one per account
CREATE TABLE datadog_accounts (
    id TEXT, -- Unique identifier of the Datadog configuration
    account_id TEXT, -- Parent account this configuration belongs to
    cost_per_gb_ingested REAL, -- Cost per GB of log data ingested (USD). NULL = using Datadog's published rate ($0.10/GB). Set to override with actual contract rate.
    created_at TEXT, -- When the Datadog account was created
    name TEXT, -- Display name for this Datadog account
    site TEXT -- Datadog regional site. Values: US1 = datadoghq.com.; US3 = us3.datadoghq.com.; US5 = us5.datadoghq.com.; EU1 = datadoghq.eu.; US1_FED = ddog-gov.com.; AP1 = ap1.datadoghq.com.; AP2 = ap2.datadoghq.com.
);

-- Discovered Datadog log index where logs are stored (e.g., main, security, compliance)
CREATE TABLE datadog_log_indexes (
    id TEXT, -- Unique identifier for this index record
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from datadog_account.account_id.
    cost_per_million_events_indexed REAL, -- Cost per million events indexed in this index (USD). NULL = using Datadog's published rate ($1.70/M). SIEM indexes cost more — set accordingly.
    created_at TEXT, -- When this index was first discovered
    datadog_account_id TEXT, -- The Datadog account this index belongs to
    name TEXT -- Index name from Datadog (e.g., 'main', 'security', 'compliance') - this is the stable identifier
);

-- Current AI curation for a finding.
CREATE TABLE finding_curations (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from finding.account_id.
    body TEXT, -- User-facing markdown-friendly explanation of the curated finding.
    created_at TEXT, -- When this curation row was created.
    disposition TEXT, -- Whether this candidate should be kept as a legitimate finding or suppressed. Values: keep = Keep the finding as a legitimate user-facing issue.; suppress = Do not keep this candidate as a user-facing finding.
    finding_id TEXT, -- The finding this curation belongs to
    finding_problem_version INTEGER, -- Problem version of the finding when this curation was last produced.
    priority TEXT, -- How much attention a kept finding deserves. Values: low = Legitimate finding, but low urgency or prominence.; medium = Legitimate finding with clear but not top-tier urgency.; high = Legitimate finding that deserves strong user attention.
    title TEXT, -- Short user-facing title for the curated finding.
    version INTEGER -- Version of the curation contract and prompt that produced this record.
);

-- Explicit membership between a finding and a log event.
CREATE TABLE finding_log_events (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from finding.account_id.
    created_at TEXT, -- When this membership row was created.
    finding_id TEXT, -- The finding this membership belongs to
    log_event_id TEXT -- The log event associated with this finding
);

-- Versioned working remediation plans for findings.
CREATE TABLE finding_plans (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from finding.account_id.
    created_at TEXT, -- When this plan row was created.
    finding_curation_id TEXT, -- The curation that produced this plan revision
    finding_id TEXT, -- The finding this plan belongs to
    -- Open questions that materially affect whether this plan should be approved or revised.
    -- JSON array of objects. Each element:
    -- $[0].scope                     string     - Scope where the missing input applies, such as service, environment, or rollout.
    -- $[0].key                       string     - Stable machine-readable key for the question.
    -- $[0].question                  string     - User-facing question that needs to be answered.
    -- $[0].why_it_matters            string     - Why the answer materially affects whether the plan should be approved or revised.
    -- $[0].choices[]                 string[]   - Suggested answer choices when the question is constrained.
    -- $[0].default_answer            string     - Suggested default answer when one exists.
    -- 
    -- Example: json_extract(open_questions, '$[0].field_name')
    open_questions TEXT,
    rationale TEXT, -- Why this plan is the preferred remediation path.
    revision INTEGER, -- Monotonic revision number for this finding's plan.
    status TEXT, -- Current lifecycle state for this plan revision. Values: draft = Initial working draft.; needs_input = Waiting on missing context or operator input.; ready_for_review = Ready for human review.; approved = Approved for execution.; executing = Execution is in progress.; completed = Execution finished successfully.; failed = Execution or validation failed.; superseded = Replaced by a newer revision.; canceled = Canceled before completion.
    -- Ordered generic remediation steps expressed as intent, not provider-specific implementation.
    -- JSON array of objects. Each element:
    -- $[0].kind                      string     - Step discriminator identifying the remediation action kind.
    -- $[0].summary                   string     - Short summary of what the remediation step does.
    -- $[0].payload                   object     - Step-specific structured payload when the step kind needs extra configuration.
    -- 
    -- Example: json_extract(steps, '$[0].field_name')
    steps TEXT,
    summary TEXT, -- Summary of the remediation plan.
    title TEXT, -- Short user-facing title for this plan.
    version INTEGER -- Version of the plan contract and prompt that produced this record.
);

-- Cache table for canonical finding_statuses view. Refreshed by worker-owned status loops.
CREATE TABLE finding_statuses_cache (
    id TEXT,
    account_id TEXT,
    current_bytes_per_hour REAL,
    current_bytes_usd_per_hour REAL,
    current_events_per_hour REAL,
    current_total_usd_per_hour REAL,
    current_volume_usd_per_hour REAL,
    finding_created_at TEXT,
    finding_id TEXT,
    finding_updated_at TEXT,
    isolated_bytes_per_hour REAL,
    isolated_bytes_usd_per_hour REAL,
    isolated_events_per_hour REAL,
    isolated_saved_bytes_per_hour REAL,
    isolated_saved_bytes_usd_per_hour REAL,
    isolated_saved_events_per_hour REAL,
    isolated_saved_total_usd_per_hour REAL,
    isolated_saved_volume_usd_per_hour REAL,
    isolated_total_usd_per_hour REAL,
    isolated_volume_usd_per_hour REAL,
    log_event_count INTEGER,
    log_event_id TEXT,
    plan_status TEXT,
    plan_updated_at TEXT,
    scope_kind TEXT, -- Explicit scope discriminator for this finding. Values: account, service, log_event.
    service_id TEXT
);

-- Durable issue instance produced by a problem over catalog facts and telemetry baselines.
CREATE TABLE findings (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Parent account this finding belongs to
    closed_at TEXT, -- When this finding stopped matching the current world. Null while active.
    created_at TEXT, -- When this finding row was created.
    -- Problem-specific typed details for this finding instance.
    -- JSON shape depends on type.
    -- 
    -- Variant: background_noise
    -- Fields:
    -- $.peer_label                   string     - Named peer label involved in the noisy interaction when one is known.
    -- $.peer_role                    string     - Coarse role of the peer involved in the noisy interaction.
    -- $.peer_kind                    string     - Coarse kind of peer involved in the noisy interaction when one is known.
    -- 
    -- Variant: commodity_traffic
    -- Fields:
    -- $.event_name                   string     - Log event name identified as commodity traffic.
    -- 
    -- Variant: dead_weight
    -- Fields:
    -- $.event_name                   string     - Log event name identified as dead weight.
    -- 
    -- Variant: debug_artifacts
    -- Fields:
    -- $.event_name                   string     - Log event name identified as a debug artifact.
    -- 
    -- Variant: debug_marker
    -- Fields:
    -- $.event_name                   string     - Log event name identified as a debug marker.
    -- 
    -- Variant: hot_path
    -- Fields:
    -- $.emission_template            string     - Stable emission template associated with the hot path event.
    -- 
    -- Variant: level_too_high
    -- Fields:
    -- $.event_name                   string     - Log event name whose level is higher than expected.
    -- $.current_level                string     - Observed log level on the event.
    -- $.expected_level               string     - Expected log level for the event.
    -- 
    -- Variant: level_too_low
    -- Fields:
    -- $.event_name                   string     - Log event name whose level is lower than expected.
    -- $.current_level                string     - Observed log level on the event.
    -- $.expected_level               string     - Expected log level for the event.
    -- 
    -- Variant: reactive_flood
    -- Fields:
    -- $.operation                    string     - Stable operation associated with the reactive flood behavior.
    -- $.outcome                      string     - Durable outcome associated with the reactive flood behavior.
    -- $.peer_label                   string     - Named peer label involved in the reactive interaction when one is known.
    -- 
    -- Variant: routine_system_chatter
    -- Fields:
    -- $.event_name                   string     - Log event name identified as routine system chatter.
    -- $.expected_level               string     - Expected log level for this routine chatter event.
    -- $.operator_prominence          string     - Operator prominence judgment associated with the event.
    -- $.instance_value               string     - Per-instance value judgment associated with the event.
    -- $.collection_gain              string     - Collection-level gain judgment associated with the event.
    -- 
    -- Variant: sensitive_data_exposure
    -- Fields:
    -- $.event_name                   string     - Log event name where the exposure was observed.
    -- $.exposure_classes             string     - Comma-separated list of exposure classes observed in the event.
    -- $.exposure_class_count         number     - Number of distinct exposure classes observed in the event.
    -- $.highest_risk_class           string     - Highest-risk exposure class observed in the event.
    -- $.has_secret_exposure          boolean    - Whether secret material was exposed.
    -- $.has_payment_exposure         boolean    - Whether payment-related data was exposed.
    -- $.has_pii_exposure             boolean    - Whether personally identifiable information was exposed.
    -- $.has_phi_exposure             boolean    - Whether protected health information was exposed.
    -- $.secret_paths                 string     - Comma-separated event paths where secret material was exposed.
    -- $.payment_paths                string     - Comma-separated event paths where payment-related data was exposed.
    -- $.pii_paths                    string     - Comma-separated event paths where personally identifiable information was exposed.
    -- $.phi_paths                    string     - Comma-separated event paths where protected health information was exposed.
    -- 
    -- Example: json_extract(details, '$.field_name') with the matching discriminator columns.
    details TEXT,
    domain TEXT, -- Top-level product domain this finding belongs to, e.g. quality or operations.
    fingerprint TEXT, -- Stable problem-specific identity for this finding instance within an account.
    log_event_id TEXT, -- Associated log event when this finding is explicitly scoped to one log event.
    problem_version INTEGER, -- Version of the problem logic that most recently reconciled this finding.
    scope_kind TEXT, -- Explicit scope discriminator for this finding. Values: account, service, log_event.
    service_id TEXT, -- Owning service for service-scoped or log-event-scoped findings. Null only for account-scoped findings.
    type TEXT -- Problem type that raised this finding, e.g. reactive_flood.
);

-- Extensible typed facts attached to a log event.
CREATE TABLE log_event_facts (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from log_event.account_id.
    created_at TEXT, -- When this fact was first recorded
    fact_name TEXT, -- Fact name within the slice, e.g. identity_profile, operational_role, value_profile.
    log_event_id TEXT, -- The log event this fact belongs to
    slice_name TEXT, -- Owned fact slice refreshed together, e.g. identity, meaning, quality.
    slice_version INTEGER, -- Current code version applied for this slice.
    -- Typed fact payload stored as JSONB. Shape is determined by slice_name + fact_name.
    -- JSON shape depends on slice_name + fact_name.
    -- 
    -- Variant: identity + identity_profile
    -- Captures the intrinsic event identity: what happened, what durable thing it is about, the durable outcome if one exists, and the stable operation when that matters.
    -- Fields:
    -- $.action                       string     - The coarse stable thing that happened.
    -- $.subject_class                string     - The coarse stable class the subject belongs to.
    -- $.subject                      string     - What specific durable thing the event is fundamentally about.
    -- $.outcome                      string     - The durable outcome, when one exists.
    -- $.operation                    string     - The stable named operation, when one exists.
    -- 
    -- Variant: attribution + attribution_profile
    -- Captures the separate materially involved peer boundary around the event when one is clearly exposed.
    -- Fields:
    -- $.peer_label                   string     - The most specific stable peer identity materially involved, if one is clear.
    -- $.peer_role                    string     - The coarse operational role of the materially involved peer. Values: user=A human or end-user actor is the materially involved peer.; service=Another application or worker service is the materially involved peer.; dependency=A supporting dependency such as a database, cache, queue, storage system, or external API is the materially involved peer.; bot=A crawler or bot actor is the materially involved peer.; probe=A health, readiness, uptime, or probe actor is the materially involved peer.; unknown=A peer is involved but the broad role is still unknown..
    -- $.peer_kind                    string     - The operational kind of the materially involved peer. Values: database=A database peer such as Postgres or MySQL.; cache=A cache peer such as Redis or Memcached.; queue=A queue or message-bus peer such as Kafka, SQS, or EventBridge.; object_store=An object storage peer such as S3 or GCS.; internal_api=A first-party internal API peer.; external_api=A third-party or otherwise external API peer.; payment_service=A payment-processing service peer such as Stripe.; email_service=An email-delivery service peer such as SES or Mailchimp.; unknown=A peer is involved but the operational kind is still unknown..
    -- $.boundary_side                string     - The coarse side of the materially involved peer boundary. Values: internal=The materially involved peer is another internal system or boundary within the same estate.; upstream=The materially involved peer appears to be upstream of the current service or is initiating work into it.; downstream=The materially involved peer appears to be downstream of the current service or is being called by it..
    -- 
    -- Variant: structural + structure_profile
    -- Captures the event-facing structural summary from log stream structure, including real stream and execution keys.
    -- Fields:
    -- $.body_kind                    string     - The deterministic body usefulness class. Values: empty=The body is empty.; generic=The body is generic and weakly descriptive.; opaque=The body is opaque or machine-encoded.; descriptive=The body is semantically descriptive..
    -- $.is_fragment                  boolean    - Whether the event is a fragment rather than a complete semantic event.
    -- $.template                     string     - The stable message or structural template when one is visible.
    -- $.stream_correlation_keys[]    string[]   - The canonical stream correlation keys observed in structure analysis.
    -- $.execution_correlation_keys[] string[]   - The canonical execution or work-unit correlation keys observed in structure analysis.
    -- $.gaps[]                       string[]   - Meaningful structural gaps observed during structure analysis. Values: empty_capture=No usable records were available for structure analysis.; missing_stream_partitions=Expected stream partitioning was not visible.; missing_execution_grouping=Expected execution grouping was not visible..
    -- 
    -- Variant: artifact + event_role_profile
    -- Classifies the event's coarse observability artifact role as signal, marker, debug artifact, or garbled output.
    -- Fields:
    -- $.role                         string     - The event's coarse observability role. Values: signal=A readable event, state change, alert, or outcome artifact.; marker=A breadcrumb, checkpoint, or lightweight progress artifact.; debug_artifact=A debug-useful payload artifact such as a request, response, dump, or snapshot emitted inline into logs.; garbled=A malformed or unintelligible artifact..
    -- 
    -- Variant: artifact + emission_template_profile
    -- Captures the stable emitted wording this event follows as an instrumentation artifact.
    -- Fields:
    -- $.template                     string     - The stable emitted wording for this event, with descriptive placeholders for clearly variable slots.
    -- 
    -- Variant: quality + operator_prominence_profile
    -- Captures how prominently the event deserves to appear in a normal operator workflow.
    -- Fields:
    -- $.operator_prominence          string     - How prominently this event deserves to appear in normal operator workflows. Values: none=Implementation-step detail or housekeeping chatter that does not deserve routine operator prominence.; low=Situationally useful proof-of-life, maintenance summary, or routine lifecycle context.; high=Important operator-facing signal about health, failures, coordination, state, or meaningful lifecycle changes..
    -- 
    -- Variant: quality + observability_value_profile
    -- Captures how much one representative event matters on its own and how much extra value the larger collection adds beyond one representative copy.
    -- Fields:
    -- $.instance_value               string     - How helpful one representative event instance is on its own. Values: none=One representative instance adds essentially no meaningful observability clue.; low=One representative instance adds a weak but real observability clue.; high=One representative instance is directly valuable on its own..
    -- $.collection_gain              string     - How much additional observability value the larger collection adds beyond one representative instance. Values: none=The larger collection adds essentially no extra value beyond one representative instance.; low=The larger collection adds limited but real extra value beyond one representative instance.; high=The larger collection adds materially important value beyond one representative instance..
    -- 
    -- Variant: quality + level_expectation_profile
    -- Captures the normalized severity level the event deserves.
    -- Fields:
    -- $.expected_level               string     - The severity level this event should use. Values: debug=Implementation-facing inspection, trace, planning, or evaluation detail.; info=Expected operator-facing work, lifecycle, housekeeping, or state-change signal.; warn=Degraded, unexpected, retrying, or risky conditions where the system is still operating and the intended work has not clearly failed outright.; error=A concrete failure or hard broken condition where the intended work did not complete..
    -- 
    -- Variant: quality + payload_hygiene_profile
    -- Captures where event-local payload weight lives across the body and event attributes, plus the exact costly paths.
    -- Fields:
    -- $.body_weight                  string     - The weight carried by the log body itself. Values: lean=The body is compact and inexpensive.; moderate=The body is moderate in size and cost.; heavy=The body is heavy and materially costlier than normal..
    -- $.attribute_weight             string     - The weight carried by event attributes. Values: lean=The event attributes are compact and inexpensive.; moderate=The event attributes are moderate in size and cost.; heavy=The event attributes are heavy and materially costlier than normal..
    -- $.expensive_paths[]            string[][] - Concrete paths that materially increase cost or clutter.
    -- 
    -- Variant: quality + context_completeness_profile
    -- Captures whether the event carries the trace and work-unit identifiers that should exist on this event, and names any concrete required paths that are missing.
    -- Fields:
    -- $.trace_context                string     - Whether the event carries the distributed trace identifiers it should have. Values: complete=Trace context is complete.; partial=Trace context is partially present.; missing=Trace context is missing.; not_applicable=Trace context does not belong on this event shape..
    -- $.work_unit_context            string     - Whether the event carries the concrete request, job, workflow, batch, message, or similar execution identifier it should have. Values: complete=Work-unit context is complete.; partial=Work-unit context is partially present.; missing=Work-unit context is missing.; not_applicable=Work-unit context does not belong on this event shape..
    -- $.missing_required_paths[]     string[][] - Concrete correlation paths that should exist on this event but are missing.
    -- 
    -- Variant: compliance + sensitive_data_profile
    -- Lists the exact event paths that actually expose secrets, payment data, PII, or PHI in sampled evidence.
    -- Fields:
    -- $.secret_paths[]               string[][] - Exact paths that expose secrets or credentials.
    -- $.payment_paths[]              string[][] - Exact paths that expose payment data.
    -- $.pii_paths[]                  string[][] - Exact paths that expose actionable personally identifying data.
    -- $.phi_paths[]                  string[][] - Exact paths that expose protected health information.
    -- 
    -- Example: json_extract(value, '$.field_name') with the matching discriminator columns.
    value TEXT
);

-- Per-log-event stream-change expressions produced from findings and merged into preview/effective states.
CREATE TABLE log_event_policies (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from log_event.account_id.
    compiled_at TEXT, -- When this policy row was last compiled from its current inputs.
    created_at TEXT, -- When this policy row was created.
    finding_id TEXT, -- Owning finding for contribution rows. Null for merged preview/effective rows.
    kind TEXT, -- Policy role for this log event. Values: contribution = One finding's contribution row.; preview = Merged preview state before enforcement.; effective = Merged enforced state currently in effect.
    log_event_id TEXT, -- The log event this policy applies to
    -- Typed compiled per-log-event stream change. Shared by contribution, preview, and effective rows.
    -- JSON object. Fields:
    -- $.drop                         object     - Drop operation applied to the event stream.
    -- $.drop.enabled                 boolean    - Whether matching events should be dropped entirely.
    -- $.sample                       object     - Sampling operation applied to the event stream.
    -- $.sample.enabled               boolean    - Whether matching events should be sampled.
    -- $.sample.interval_seconds      number     - Sampling interval in seconds when sampling is enabled.
    -- $.trim                         object     - Trim operation that removes expensive fields from the event payload.
    -- $.trim.enabled                 boolean    - Whether matching events should have fields trimmed.
    -- $.trim.target_paths[]          string[][] - Event field paths that should be removed from the payload.
    -- $.redact                       object     - Redaction operations applied to sensitive fields in the event payload.
    -- $.redact.entries[]             object[]   - Redaction entries applied to the payload.
    -- $.redact.entries[].target_paths[] string[][] - Event field paths that should be redacted.
    -- $.rewrite                      object     - Rewrite operations that change severity or move fields within the event payload.
    -- $.rewrite.severity             object     - Severity rewrite applied to the event.
    -- $.rewrite.severity.value       string     - New severity value assigned to the event.
    -- $.rewrite.fields[]             object[]   - Field move rewrites applied to the event payload.
    -- $.rewrite.fields[].from[]      string[]   - Source field path in the event payload.
    -- $.rewrite.fields[].to[]        string[]   - Destination field path in the event payload.
    -- 
    -- Example: json_extract(spec, '$.field_name')
    spec TEXT
);

-- Cache table for canonical log_event_statuses view. Refreshed by worker-owned status loops.
CREATE TABLE log_event_statuses_cache (
    id TEXT,
    account_id TEXT,
    current_bytes_per_hour REAL,
    current_bytes_usd_per_hour REAL,
    current_events_per_hour REAL,
    current_total_usd_per_hour REAL,
    current_volume_usd_per_hour REAL,
    effective_bytes_per_hour REAL,
    effective_bytes_usd_per_hour REAL,
    effective_events_per_hour REAL,
    effective_saved_bytes_per_hour REAL,
    effective_saved_bytes_usd_per_hour REAL,
    effective_saved_events_per_hour REAL,
    effective_saved_total_usd_per_hour REAL,
    effective_saved_volume_usd_per_hour REAL,
    effective_total_usd_per_hour REAL,
    effective_volume_usd_per_hour REAL,
    has_been_analyzed INTEGER,
    has_effective_policy INTEGER,
    has_preview_policy INTEGER,
    has_volumes INTEGER,
    log_event_id TEXT,
    preview_bytes_per_hour REAL,
    preview_bytes_usd_per_hour REAL,
    preview_events_per_hour REAL,
    preview_saved_bytes_per_hour REAL,
    preview_saved_bytes_usd_per_hour REAL,
    preview_saved_events_per_hour REAL,
    preview_saved_total_usd_per_hour REAL,
    preview_saved_volume_usd_per_hour REAL,
    preview_total_usd_per_hour REAL,
    preview_volume_usd_per_hour REAL,
    service_id TEXT
);

-- Distinct log message pattern discovered within a service. Defines how to parse and match logs using codecs and matchers.
CREATE TABLE log_events (
    id TEXT, -- Unique identifier of the log event
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from service.account_id.
    baseline_avg_bytes REAL, -- Current trailing 7-day volume-weighted average bytes/event. Refreshed on volume ingestion.
    baseline_volume_per_hour REAL, -- Current trailing 7-day average events/hour. Refreshed on volume ingestion.
    created_at TEXT, -- When the log event was created
    description TEXT, -- What the event is and what data instances carry. Helps engineers decide whether to look here.
    -- JSON rules that match incoming logs to this event. Each matcher specifies a field path, operator, and value.
    -- JSON array of objects. Each element:
    -- $[0].field_path[]              string[]   - Path to field as array of segments
    -- $[0].match_type                string     - Match operator: exact, contains, starts_with, ends_with, regex, exists, missing
    -- $[0].match_value               string     - Value to match against
    -- $[0].case_insensitive          boolean    - Whether matching is case-insensitive (optional)
    -- $[0].negate                    boolean    - Whether to invert match result (optional)
    -- 
    -- Example: json_extract(matchers, '$[0].field_name')
    matchers TEXT,
    name TEXT, -- Snake_case identifier unique per service, e.g. nginx_access_log
    service_id TEXT, -- Service that produces this event
    severity TEXT -- Predominant log severity level, derived from example records. Nullable when examples have no severity info. Values: debug = The event usually appears at debug severity.; info = The event usually appears at info severity.; warn = The event usually appears at warn severity.; error = The event usually appears at error severity.; other = The event usually appears at a non-standard or other severity.
);

-- Single message in a chat conversation. Append-only — never updated or deleted.
CREATE TABLE messages (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from conversation.account_id.
    -- Array of typed content blocks: text, thinking, tool_use, tool_result
    -- JSON array of objects. Each element:
    -- $[0].type                      string     - Block type discriminator: text, thinking, tool_use, tool_result
    -- $[0].text                      object     - Text content (when type=text)
    -- $[0].text.content              string     - Text content
    -- $[0].thinking                  object     - AI reasoning content (when type=thinking)
    -- $[0].thinking.content          string     - AI reasoning content
    -- $[0].tool_use                  object     - Tool call (when type=tool_use)
    -- $[0].tool_use.id               string     - Unique tool call identifier
    -- $[0].tool_use.name             string     - Tool name
    -- $[0].tool_use.input            object     - Tool input parameters as JSON object
    -- $[0].tool_result               object     - Tool response (when type=tool_result)
    -- $[0].tool_result.tool_use_id   string     - ID of the tool use this result corresponds to
    -- $[0].tool_result.is_error      boolean    - Whether the tool call failed
    -- $[0].tool_result.error         string     - Human-readable error message (when is_error=true)
    -- $[0].tool_result.content       object     - Structured result data (when is_error=false)
    -- 
    -- Example: json_extract(content, '$[0].field_name')
    content TEXT,
    conversation_id TEXT, -- Conversation this message belongs to
    created_at TEXT, -- When the message was created
    model TEXT, -- AI model that produced this message. Null for user messages.
    role TEXT, -- Who sent this message. Values: user = Human-originated message content.; assistant = AI-originated message content.
    stop_reason TEXT -- Why the assistant stopped generating. Null for user messages. Values: end_turn = The assistant completed its response.; tool_use = The assistant paused to call a tool.
);

-- Extensible typed facts attached to a service.
CREATE TABLE service_facts (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from service.account_id.
    created_at TEXT, -- When this fact was first recorded
    fact_group TEXT, -- Owned fact group refreshed together, e.g. catalog_service_facts.
    fact_type TEXT, -- Fact type within the namespace, e.g. service_profile or actionability_profile.
    namespace TEXT, -- Fact namespace, e.g. semantic or operational.
    service_id TEXT, -- The service this fact belongs to
    -- Typed fact payload stored as JSONB. Shape is determined by namespace + fact_type.
    -- JSON shape depends on namespace + fact_type.
    -- 
    -- Variant: semantic + service_profile
    -- Describes what the service appears to be, what it is primarily responsible for, and the major role it plays in the system.
    -- Fields:
    -- $.summary                      string     - A concise grounded summary of what the service appears to do.
    -- $.service_category             string     - The best-fitting broad service category. Values: customer_api=Customer-facing request/response API service.; internal_api=Internal API or control-plane service.; background_worker=Asynchronous worker or job processor.; integration_adapter=Connector, bridge, or external integration adapter.; data_pipeline=Pipeline, stream processor, or ETL-style service.; platform_component=Platform or infrastructure component.; message_broker=Messaging broker or queueing component.; database=Database or durable storage engine.; cache=Cache or in-memory data service.; observability_component=Collector, agent, or observability infrastructure.; unknown=Insufficient evidence for a stronger classification..
    -- $.primary_responsibilities[]   string[]   - Short grounded statements of the service's primary responsibilities.
    -- $.system_roles[]               string[]   - Major system roles the service appears to play. Values: request_handling=Handles requests or API traffic.; background_processing=Processes asynchronous or scheduled work.; integration=Connects to or mediates external systems.; state_management=Stores or manages application state.; messaging=Publishes, consumes, or routes messages.; platform_operations=Performs platform or infrastructure operations..
    -- 
    -- Variant: semantic + telemetry_profile
    -- Describes the broad shape of telemetry this service emits, including the main log patterns and notable telemetry characteristics.
    -- Fields:
    -- $.summary                      string     - A concise grounded summary of what the service's logs mostly look like.
    -- $.dominant_log_patterns[]      string[]   - The main classes of logs the service appears to emit. Values: request_logs=Request or access logs.; job_logs=Background job or worker logs.; integration_logs=External-system integration logs.; lifecycle_logs=Startup, shutdown, and lifecycle logs.; platform_housekeeping_logs=Platform or housekeeping logs.; audit_logs=Audit or compliance-oriented logs..
    -- $.dominant_operational_roles[] string[]   - Plain-language descriptions of the main operational roles visible in the logs.
    -- $.telemetry_characteristics[]  string[]   - Broad traits of the service's telemetry. Values: high_volume=High log volume relative to the visible behaviors.; error_heavy=Errors are a large visible part of the stream.; infra_heavy=Infrastructure or platform logs dominate the stream.; integration_heavy=Integration activity dominates the stream.; state_transition_heavy=State-change or workflow logs dominate the stream.; mixed_signal_quality=The stream mixes high-value and low-value signal quality..
    -- 
    -- Variant: operational + actionability_profile
    -- Describes how much direct control the team likely has over this service and which change levers are most realistic when a finding affects it.
    -- Fields:
    -- $.summary                      string     - A concise grounded summary of how actionable findings against this service are likely to be.
    -- $.service_origin               string     - Where this service likely comes from. Values: first_party=Application or service written and owned by the team.; third_party=Third-party product or vendor-managed service.; open_source_component=Open-source component the team operates.; managed_platform=Managed platform or infrastructure service.; unknown=Insufficient evidence for a stronger classification..
    -- $.ownership_level              string     - How much direct control the team likely has over behavior changes. Values: owned_code=The team can directly change the service code.; configured_only=The team mostly influences behavior through configuration.; operated_only=The team mainly operates the service but does not own its code.; unknown=Insufficient evidence for a stronger classification..
    -- $.change_levers[]              string[]   - The most realistic levers available for changing this service or its telemetry. Values: application_code=Application code changes.; service_config=Service-level configuration changes.; deployment_config=Deployment or infrastructure-as-code changes.; collector_config=Collector or telemetry-pipeline configuration changes.; platform_tuning=Platform tuning or operational settings..
    -- 
    -- Variant: operational + shared_payload_hygiene_profile
    -- Describes the overall weight and non-essential noise of shared service-level telemetry metadata, along with the exact expensive shared paths.
    -- Fields:
    -- $.shared_payload_weight        string     - The overall payload weight of the shared service-level metadata. Values: lean=Shared metadata is compact and inexpensive.; moderate=Shared metadata is noticeable but not dominating.; heavy=Shared metadata is heavy and materially increases payload cost..
    -- $.shared_enrichment_noise      string     - How much non-essential shared enrichment surrounds the service's logs. Values: none=There is no meaningful shared enrichment noise.; low=There is limited shared enrichment noise.; medium=There is noticeable shared enrichment noise.; high=There is heavy shared enrichment noise..
    -- $.expensive_shared_paths[]     string[][] - Concrete shared paths that materially increase cost or clutter across this service's logs.
    -- 
    -- Example: json_extract(value, '$.field_name') with the matching discriminator columns.
    value TEXT,
    version INTEGER -- Current code version applied for this fact group.
);

-- Cache table for canonical service_statuses view. Refreshed by worker-owned status loops.
CREATE TABLE service_statuses_cache (
    id TEXT,
    account_id TEXT,
    current_bytes_per_hour REAL,
    current_bytes_usd_per_hour REAL,
    current_events_per_hour REAL,
    current_service_debug_events_per_hour REAL,
    current_service_error_events_per_hour REAL,
    current_service_events_per_hour REAL,
    current_service_info_events_per_hour REAL,
    current_service_other_events_per_hour REAL,
    current_service_volume_usd_per_hour REAL,
    current_service_warn_events_per_hour REAL,
    current_total_usd_per_hour REAL,
    current_volume_usd_per_hour REAL,
    datadog_account_id TEXT,
    effective_bytes_per_hour REAL,
    effective_bytes_usd_per_hour REAL,
    effective_events_per_hour REAL,
    effective_log_event_count INTEGER,
    effective_saved_bytes_per_hour REAL,
    effective_saved_bytes_usd_per_hour REAL,
    effective_saved_events_per_hour REAL,
    effective_saved_total_usd_per_hour REAL,
    effective_saved_volume_usd_per_hour REAL,
    effective_total_usd_per_hour REAL,
    effective_volume_usd_per_hour REAL,
    health TEXT, -- Health state for this status row. Values: DISABLED = The user turned the log pipeline off.; INACTIVE = The service currently has no log volume.; ERROR = The log pipeline is unhealthy or failing.; OK = The log pipeline is healthy.
    log_event_analyzed_count INTEGER,
    log_event_count INTEGER,
    preview_bytes_per_hour REAL,
    preview_bytes_usd_per_hour REAL,
    preview_events_per_hour REAL,
    preview_log_event_count INTEGER,
    preview_saved_bytes_per_hour REAL,
    preview_saved_bytes_usd_per_hour REAL,
    preview_saved_events_per_hour REAL,
    preview_saved_total_usd_per_hour REAL,
    preview_saved_volume_usd_per_hour REAL,
    preview_total_usd_per_hour REAL,
    preview_volume_usd_per_hour REAL,
    service_id TEXT
);

-- Maps a service to its owning team.
CREATE TABLE service_team_mappings (
    id TEXT, -- Unique identifier of the service-team mapping
    account_id TEXT, -- Denormalized account for RLS. Auto-set via trigger from service.account_id.
    created_at TEXT, -- When the mapping was created
    service_id TEXT, -- Service assigned to a team
    team_id TEXT, -- Owning team for the service
    updated_at TEXT -- When the mapping was last updated
);

-- Application or microservice that produces logs. Central entity in the data catalog.
CREATE TABLE services (
    id TEXT, -- Unique identifier of the service
    account_id TEXT, -- Parent account this service belongs to
    created_at TEXT, -- When the service was created
    enabled INTEGER, -- Whether log analysis and policy generation is active for this service
    initial_weekly_log_count INTEGER, -- Approximate weekly log count from initial catalog loop pass (7-day period from Datadog)
    name TEXT -- Service identifier in telemetry (e.g., 'checkout-service')
);

-- Membership of a WorkOS user in an organization-scoped team.
CREATE TABLE team_memberships (
    id TEXT, -- Unique identifier of the team membership
    created_at TEXT, -- When the membership was created
    organization_id TEXT, -- Denormalized organization for RLS. Auto-set via trigger from team.organization_id.
    team_id TEXT, -- Team this membership belongs to
    updated_at TEXT, -- When the membership was last updated
    user_id TEXT -- WorkOS user ID assigned to the team
);

-- Organization-scoped group of users used for service ownership and filtering.
CREATE TABLE teams (
    id TEXT, -- Unique identifier of the team
    created_at TEXT, -- When the team was created
    description TEXT, -- Optional description of the team's scope and responsibilities
    external_id TEXT, -- Optional external group identifier reserved for future SCIM mapping
    name TEXT, -- Human-readable team name within the organization
    organization_id TEXT, -- Organization this team belongs to
    updated_at TEXT -- When the team was last updated
);

-- Purpose-aligned environment for reviewing and classifying telemetry.
CREATE TABLE workspaces (
    id TEXT, -- Unique identifier of the workspace
    account_id TEXT, -- Parent account this workspace belongs to
    created_at TEXT, -- When the workspace was created
    name TEXT, -- Human-readable name within the account
    purpose TEXT -- Primary purpose determining evaluation strategy. Values: observability = Performance, reliability, and operational visibility.; security = Threat detection, investigation, and security posture.; compliance = Regulatory, privacy, and policy compliance review.
);

