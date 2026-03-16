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

-- Live working set of entities (services, log events) referenced in a conversation
CREATE TABLE conversation_contexts (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from conversation.account_id.
    added_by TEXT, -- Who added this entity to context. user: added via @-reference, assistant: added by AI during chat.
    conversation_id TEXT, -- Conversation this context belongs to
    created_at TEXT, -- When the entity was added to context
    entity_id TEXT, -- ID of the context entity
    entity_type TEXT -- Type of the context entity. service: an application producing logs, log_event: a specific event pattern.
);

-- Chat session between a user and the AI assistant within a workspace
CREATE TABLE conversations (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from workspace.account_id.
    created_at TEXT, -- When the conversation was created
    title TEXT, -- AI-generated title, set after first exchange
    user_id TEXT, -- WorkOS user ID who owns this conversation
    view_id TEXT, -- If set, this conversation is for iterating on a specific view
    workspace_id TEXT -- Workspace this conversation belongs to
);

-- Cache table for canonical datadog_account_statuses view. Refreshed by worker-owned status loops.
CREATE TABLE datadog_account_statuses_cache (
    id TEXT,
    account_id TEXT,
    approved_recommendation_count INTEGER,
    current_bytes_per_hour REAL,
    current_bytes_usd_per_hour REAL,
    current_events_per_hour REAL,
    current_service_events_per_hour REAL,
    current_service_volume_usd_per_hour REAL,
    current_total_usd_per_hour REAL,
    current_volume_usd_per_hour REAL,
    datadog_account_id TEXT,
    disabled_services INTEGER,
    dismissed_recommendation_count INTEGER,
    estimated_bytes_per_hour REAL,
    estimated_bytes_usd_per_hour REAL,
    estimated_events_per_hour REAL,
    estimated_total_usd_per_hour REAL,
    estimated_volume_usd_per_hour REAL,
    health TEXT, -- Values: DISABLED, INACTIVE, ERROR, OK.
    impact_bytes_per_hour REAL,
    impact_bytes_usd_per_hour REAL,
    impact_events_per_hour REAL,
    impact_total_usd_per_hour REAL,
    impact_volume_usd_per_hour REAL,
    inactive_services INTEGER,
    log_active_services INTEGER,
    log_event_analyzed_count INTEGER,
    log_event_count INTEGER,
    log_service_count INTEGER,
    ok_services INTEGER,
    pending_recommendation_count INTEGER,
    policy_pending_critical_count INTEGER,
    policy_pending_high_count INTEGER,
    policy_pending_low_count INTEGER,
    policy_pending_medium_count INTEGER,
    ready_for_use INTEGER,
    recommendation_count INTEGER
);

-- Datadog integration configuration for an account, one per account
CREATE TABLE datadog_accounts (
    id TEXT, -- Unique identifier of the Datadog configuration
    account_id TEXT, -- Parent account this configuration belongs to
    cost_per_gb_ingested REAL, -- Cost per GB of log data ingested (USD). NULL = using Datadog's published rate ($0.10/GB). Set to override with actual contract rate.
    created_at TEXT, -- When the Datadog account was created
    name TEXT, -- Display name for this Datadog account
    site TEXT -- Datadog regional site. US1: datadoghq.com, US3: us3.datadoghq.com, US5: us5.datadoghq.com, EU1: datadoghq.eu, US1_FED: ddog-gov.com, AP1: ap1.datadoghq.com, AP2: ap2.datadoghq.com.
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

-- AI-generated recommendation for a specific quality category on a log event, scoped to a workspace for approval
CREATE TABLE log_event_recommendations (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from workspace.account_id.
    action TEXT, -- What this policy does when enforced: 'drop' (remove all events), 'sample' (keep at reduced rate), 'filter' (drop subset by field value), 'trim' (remove/truncate fields), 'none' (informational only). Auto-set via trigger.
    -- Category-specific analysis from AI. JSON object with one field populated matching the category, containing the analysis and recommended actions.
    -- JSON object. Fields:
    -- $.pii_leakage                  - PII leakage analysis (optional)
    -- $.pii_leakage.fields[]         - List of fields that may contain PII
    -- $.pii_leakage.fields[].path[]  string[]   - Path to field as array of segments
    -- $.pii_leakage.fields[].types[] string[]   - List of sensitive data types this field may contain
    -- $.pii_leakage.fields[].observed boolean    - Whether actual sensitive data was seen in examples
    -- $.secrets_leakage              - Secrets leakage analysis (optional)
    -- $.secrets_leakage.fields[]     - List of fields that may contain secrets
    -- $.phi_leakage                  - PHI leakage analysis (optional)
    -- $.phi_leakage.fields[]         - List of fields that may contain PHI
    -- $.payment_data_leakage         - Payment data leakage analysis (optional)
    -- $.payment_data_leakage.fields[] - List of fields that may contain payment data
    -- $.health_checks                - Health checks analysis (optional)
    -- $.bot_traffic                  - Bot traffic analysis (optional)
    -- $.bot_traffic.user_agent_field[] string[]   - Path to user-agent field as array of segments
    -- $.bot_traffic.bot_proportion   number     - Fraction of traffic identified as bot/crawler (optional)
    -- $.debug_artifacts              - Debug artifacts analysis (optional)
    -- $.malformed                    - Malformed data analysis (optional)
    -- $.broken_records               - Broken records analysis (optional)
    -- $.broken_records.min_interval_seconds number     - Suggested minimum interval between kept events in seconds
    -- $.commodity_traffic            - Commodity traffic analysis (optional)
    -- $.commodity_traffic.min_interval_seconds number     - Suggested minimum interval between kept events in seconds
    -- $.redundant_events             - Redundant events analysis (optional)
    -- $.dead_weight                  - Dead weight analysis (optional)
    -- $.duplicate_fields             - Duplicate fields analysis (optional)
    -- $.duplicate_fields.pairs[]     - List of duplicate field pairs
    -- $.duplicate_fields.pairs[].remove[] string[][] - List of duplicate field paths to remove
    -- $.duplicate_fields.pairs[].keep[] string[]   - Canonical field path to keep
    -- $.instrumentation_bloat        - Instrumentation bloat analysis (optional)
    -- $.instrumentation_bloat.fields[] string[][] - List of field paths that are instrumentation bloat
    -- $.oversized_fields             - Oversized fields analysis (optional)
    -- $.oversized_fields.fields[]    string[][] - List of field paths that are oversized
    -- $.wrong_level                  - Wrong level analysis (optional)
    -- $.wrong_level.current_level    string     - Current normalized severity level
    -- $.wrong_level.suggested_level  string     - Suggested normalized severity level
    -- 
    -- Example: json_extract(analysis, '$.field_name')
    analysis TEXT,
    approved_at TEXT, -- When this policy was approved by a user
    approved_baseline_avg_bytes REAL, -- Baseline avg bytes frozen at approval time. Snapshot of log_event.baseline_avg_bytes.
    approved_baseline_volume_per_hour REAL, -- Baseline volume/hour frozen at approval time. Snapshot of log_event.baseline_volume_per_hour.
    approved_by TEXT, -- User ID who approved this policy
    category TEXT, -- Quality issue category this policy addresses. Compliance: pii_leakage, secrets_leakage, phi_leakage, payment_data_leakage. Waste: health_checks, bot_traffic, debug_artifacts, malformed, broken_records, commodity_traffic, redundant_events, dead_weight. Quality: duplicate_fields, instrumentation_bloat, oversized_fields, wrong_level.
    category_type TEXT, -- Type of problem: compliance (legal/security risk), waste (event-level cuts), or quality (field-level improvements). Auto-set via trigger from CategoryMeta.
    created_at TEXT, -- When this policy was created
    dismissed_at TEXT, -- When this policy was dismissed by a user
    dismissed_by TEXT, -- User ID who dismissed this policy
    log_event_id TEXT, -- The log event this policy applies to
    severity TEXT, -- Max compliance severity across sensitivity types. NULL for non-compliance categories. Auto-set via trigger. Values: low, medium, high, critical.
    subjective INTEGER, -- Whether this category requires AI judgment (true) vs mechanically verifiable (false). Auto-set via trigger from CategoryMeta.
    workspace_id TEXT -- The workspace that owns this policy
);

-- Cache table for canonical log_event_statuses view. Refreshed by worker-owned status loops.
CREATE TABLE log_event_statuses_cache (
    id TEXT,
    account_id TEXT,
    approved_recommendation_count INTEGER,
    current_bytes_per_hour REAL,
    current_bytes_usd_per_hour REAL,
    current_events_per_hour REAL,
    current_total_usd_per_hour REAL,
    current_volume_usd_per_hour REAL,
    dismissed_recommendation_count INTEGER,
    effective_policy_enabled INTEGER,
    estimated_bytes_per_hour REAL,
    estimated_bytes_usd_per_hour REAL,
    estimated_events_per_hour REAL,
    estimated_total_usd_per_hour REAL,
    estimated_volume_usd_per_hour REAL,
    has_been_analyzed INTEGER,
    has_volumes INTEGER,
    impact_bytes_per_hour REAL,
    impact_bytes_usd_per_hour REAL,
    impact_events_per_hour REAL,
    impact_total_usd_per_hour REAL,
    impact_volume_usd_per_hour REAL,
    log_event_id TEXT,
    pending_recommendation_count INTEGER,
    policy_pending_critical_count INTEGER,
    policy_pending_high_count INTEGER,
    policy_pending_low_count INTEGER,
    policy_pending_medium_count INTEGER,
    recommendation_count INTEGER,
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
    -- Sample log records captured during catalog loop collection, used for AI inference and pattern validation.
    -- JSON array of objects. Each element:
    -- $[0].timestamp                 - When the log event occurred (RFC3339)
    -- $[0].body                      string     - Log message content
    -- $[0].severity_text             string     - Severity level as text (e.g., INFO, ERROR)
    -- $[0].severity_number           number     - OTel severity level number (1-24)
    -- $[0].trace_id                  string     - Distributed trace ID (optional)
    -- $[0].span_id                   string     - Span ID within trace (optional)
    -- $[0].attributes                object     - Log-level attributes (http.status, error.message, etc.)
    -- $[0].resource_attributes       object     - Resource attributes (service.name, deployment.environment, etc.)
    -- $[0].scope_attributes          object     - Instrumentation scope attributes (optional)
    -- 
    -- Example: json_extract(examples, '$[0].field_name')
    examples TEXT,
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
    severity TEXT -- Predominant log severity level, derived from example records. Nullable when examples have no severity info. Values: debug, info, warn, error, other.
);

-- Single message in a chat conversation. Append-only — never updated or deleted.
CREATE TABLE messages (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from conversation.account_id.
    -- Array of typed content blocks: text, thinking, tool_use, tool_result
    -- JSON array of objects. Each element:
    -- $[0].type                      string     - Block type discriminator: text, thinking, tool_use, tool_result
    -- $[0].text                      - Text content (when type=text)
    -- $[0].text.content              string     - Text content
    -- $[0].thinking                  - AI reasoning content (when type=thinking)
    -- $[0].thinking.content          string     - AI reasoning content
    -- $[0].tool_use                  - Tool call (when type=tool_use)
    -- $[0].tool_use.id               string     - Unique tool call identifier
    -- $[0].tool_use.name             string     - Tool name
    -- $[0].tool_use.input[]          number[]   - Tool input parameters as JSON object
    -- $[0].tool_result               - Tool response (when type=tool_result)
    -- $[0].tool_result.tool_use_id   string     - ID of the tool use this result corresponds to
    -- $[0].tool_result.is_error      boolean    - Whether the tool call failed
    -- $[0].tool_result.error         string     - Human-readable error message (when is_error=true)
    -- $[0].tool_result.content[]     number[]   - Structured result data (when is_error=false)
    -- 
    -- Example: json_extract(content, '$[0].field_name')
    content TEXT,
    conversation_id TEXT, -- Conversation this message belongs to
    created_at TEXT, -- When the message was created
    model TEXT, -- AI model that produced this message. Null for user messages.
    role TEXT, -- Who sent this message. user: human-originated, assistant: AI-originated.
    stop_reason TEXT -- Why the assistant stopped generating. end_turn: completed response, tool_use: paused to call a tool. Null for user messages.
);

-- Cache table for per-category recommendation aggregations. Refreshed by worker-owned status loops.
CREATE TABLE recommendation_category_statuses_cache (
    id TEXT,
    account_id TEXT, -- Account ID for tenant isolation
    action TEXT, -- What the policy does: drop (remove events), sample (reduce rate), filter (drop subset), trim (modify fields), none (informational)
    approved_count INTEGER, -- Policies approved by user in this category
    boundary TEXT, -- Where this category stops applying — what NOT to flag
    category TEXT, -- Quality issue category (e.g., pii_leakage, noise, health_checks)
    category_type TEXT, -- Type of problem: compliance (legal/security risk), waste (event-level cuts), quality (field-level improvements).
    dismissed_count INTEGER, -- Policies dismissed by user in this category
    display_name TEXT, -- Human-readable category name (e.g., 'PII Leakage')
    estimated_bytes_reduction_per_hour REAL, -- Bytes/hour saved by all pending policies in this category combined
    estimated_cost_reduction_per_hour_bytes_usd REAL, -- Estimated ingestion savings in USD/hour from pending policies in this category
    estimated_cost_reduction_per_hour_usd REAL, -- Estimated total savings in USD/hour from pending policies in this category
    estimated_cost_reduction_per_hour_volume_usd REAL, -- Estimated indexing savings in USD/hour from pending policies in this category
    estimated_volume_reduction_per_hour REAL, -- Events/hour saved by all pending policies in this category combined
    events_with_volumes INTEGER, -- Log events in this category that have volume data (subset of total_event_count)
    pending_count INTEGER, -- Policies awaiting user review in this category
    policy_pending_critical_count INTEGER, -- Pending policies with critical compliance severity
    policy_pending_high_count INTEGER, -- Pending policies with high compliance severity
    policy_pending_low_count INTEGER, -- Pending policies with low compliance severity
    policy_pending_medium_count INTEGER, -- Pending policies with medium compliance severity
    principle TEXT, -- What this category detects — the fundamental test for membership
    subjective INTEGER, -- Whether this category requires AI judgment (true) vs mechanically verifiable (false)
    total_event_count INTEGER -- Total log events that have a policy in this category
);

-- Cache table for recommendation_statuses view. Refreshed by worker-owned status loops.
CREATE TABLE recommendation_statuses_cache (
    id TEXT,
    account_id TEXT, -- Account ID for tenant isolation
    action TEXT, -- What the policy does: drop (remove events), sample (reduce rate), filter (drop subset), trim (modify fields), none (informational)
    approved_at TEXT, -- When this policy was approved by a user
    category TEXT, -- Quality issue category this policy addresses (e.g., pii_leakage, noise, health_checks)
    category_type TEXT, -- Type of problem: compliance (legal/security risk), waste (event-level cuts), quality (field-level improvements).
    created_at TEXT, -- When this policy was created
    current_bytes_per_hour REAL, -- Current bytes/hour baseline for this recommendation
    current_bytes_usd_per_hour REAL, -- Current bytes-billed USD/hour baseline
    current_events_per_hour REAL, -- Current events/hour baseline for this recommendation
    current_total_usd_per_hour REAL, -- Current total USD/hour baseline
    current_volume_usd_per_hour REAL, -- Current volume-billed USD/hour baseline
    dismissed_at TEXT, -- When this policy was dismissed by a user
    estimated_bytes_per_hour REAL, -- Estimated bytes/hour if this recommendation is applied alone
    estimated_bytes_usd_per_hour REAL, -- Estimated bytes-billed USD/hour if applied alone
    estimated_events_per_hour REAL, -- Estimated events/hour if this recommendation is applied alone
    estimated_total_usd_per_hour REAL, -- Estimated total USD/hour if applied alone
    estimated_volume_usd_per_hour REAL, -- Estimated volume-billed USD/hour if applied alone
    impact_bytes_per_hour REAL, -- Estimated impact in bytes/hour (estimated - current)
    impact_bytes_usd_per_hour REAL, -- Estimated impact in bytes-billed USD/hour (estimated - current)
    impact_events_per_hour REAL, -- Estimated impact in events/hour (estimated - current)
    impact_total_usd_per_hour REAL, -- Estimated impact in total USD/hour (estimated - current)
    impact_volume_usd_per_hour REAL, -- Estimated impact in volume-billed USD/hour (estimated - current)
    log_event_id TEXT, -- The log event this policy targets
    log_event_name TEXT, -- Name of the targeted log event (denormalized for display)
    recommendation_id TEXT, -- The recommendation this status row represents
    service_id TEXT, -- Service that produces the targeted log event (denormalized)
    service_name TEXT, -- Name of the service (denormalized for display)
    severity TEXT, -- Max compliance severity across sensitivity types. NULL for non-compliance categories. Values: low, medium, high, critical.
    status TEXT, -- User decision on this policy. PENDING (awaiting review), APPROVED (accepted for enforcement), DISMISSED (rejected by user).
    subjective INTEGER, -- Whether this category requires AI judgment (true) vs mechanically verifiable (false)
    survival_rate REAL, -- Fraction of events that survive this policy (0.0 = all dropped, 1.0 = all kept). NULL if not estimable.
    workspace_id TEXT -- The workspace that owns this policy
);

-- Cache table for canonical service_statuses view. Refreshed by worker-owned status loops.
CREATE TABLE service_statuses_cache (
    id TEXT,
    account_id TEXT,
    approved_recommendation_count INTEGER,
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
    dismissed_recommendation_count INTEGER,
    estimated_bytes_per_hour REAL,
    estimated_bytes_usd_per_hour REAL,
    estimated_events_per_hour REAL,
    estimated_total_usd_per_hour REAL,
    estimated_volume_usd_per_hour REAL,
    health TEXT, -- Values: DISABLED, INACTIVE, ERROR, OK.
    impact_bytes_per_hour REAL,
    impact_bytes_usd_per_hour REAL,
    impact_events_per_hour REAL,
    impact_total_usd_per_hour REAL,
    impact_volume_usd_per_hour REAL,
    log_event_analyzed_count INTEGER,
    log_event_count INTEGER,
    pending_recommendation_count INTEGER,
    policy_pending_critical_count INTEGER,
    policy_pending_high_count INTEGER,
    policy_pending_low_count INTEGER,
    policy_pending_medium_count INTEGER,
    recommendation_count INTEGER,
    service_id TEXT
);

-- Application or microservice that produces logs. Central entity in the data catalog.
CREATE TABLE services (
    id TEXT, -- Unique identifier of the service
    account_id TEXT, -- Parent account this service belongs to
    created_at TEXT, -- When the service was created
    description TEXT, -- AI-generated description of what this service does and its telemetry characteristics
    enabled INTEGER, -- Whether log analysis and policy generation is active for this service
    initial_weekly_log_count INTEGER, -- Approximate weekly log count from initial catalog loop pass (7-day period from Datadog)
    name TEXT -- Service identifier in telemetry (e.g., 'checkout-service')
);

-- Group of users within a workspace that reviews policies and manages services
CREATE TABLE teams (
    id TEXT, -- Unique identifier of the team
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from workspace.account_id.
    created_at TEXT, -- When the team was created
    name TEXT, -- Human-readable name within the workspace
    workspace_id TEXT -- Parent workspace this team belongs to
);

-- User's saved reference to a view, elevating it from conversation history to their personal collection
CREATE TABLE view_favorites (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from view.account_id.
    created_at TEXT, -- When the view was favorited
    user_id TEXT, -- WorkOS user ID who favorited this view
    view_id TEXT -- The view being favorited
);

-- Saved SQL query against the local catalog, created by the AI assistant. Immutable — editing creates a fork.
CREATE TABLE views (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from message.account_id.
    conversation_id TEXT, -- Denormalized from message for easier queries
    created_at TEXT, -- When the view was created
    created_by TEXT, -- WorkOS user ID who triggered this view creation
    entity_type TEXT, -- Which catalog entity this view queries. service: applications, log_event: event patterns, policy: quality recommendations.
    forked_from_id TEXT, -- Parent view if this is a refinement/iteration
    message_id TEXT, -- Assistant message that created this view via show_view tool call
    query TEXT -- Raw SQL query executed against the client's local SQLite database
);

-- Purpose-aligned environment for reviewing and classifying telemetry. Each workspace has its own policies and teams.
CREATE TABLE workspaces (
    id TEXT, -- Unique identifier of the workspace
    account_id TEXT, -- Parent account this workspace belongs to
    created_at TEXT, -- When the workspace was created
    name TEXT, -- Human-readable name within the account
    purpose TEXT -- Primary purpose determining evaluation strategy. observability: performance and reliability, security: threat detection, compliance: regulatory requirements.
);

