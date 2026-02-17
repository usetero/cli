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
    updated_at TEXT, -- When the conversation was last updated
    user_id TEXT, -- WorkOS user ID who owns this conversation
    view_id TEXT, -- If set, this conversation is for iterating on a specific view
    workspace_id TEXT -- Workspace this conversation belongs to
);

-- Cache table for datadog_account_statuses view. Refreshed by cron service.
CREATE TABLE datadog_account_statuses_cache (
    id TEXT,
    account_id TEXT, -- Account ID (denormalized from Datadog account)
    datadog_account_id TEXT, -- The Datadog account this status belongs to
    disabled_services INTEGER, -- Services with DISABLED health
    error TEXT, -- Most recent error message from any service with ERROR health or account-level discovery
    error_at TEXT, -- When the most recent error occurred
    error_services INTEGER, -- Services with ERROR health
    estimated_bytes_reduction_per_hour REAL, -- Account-wide estimated bytes reduction
    estimated_cost_reduction_per_hour_bytes_usd REAL, -- Account-wide estimated bytes-based USD/hour savings
    estimated_cost_reduction_per_hour_usd REAL, -- Account-wide estimated total USD/hour savings
    estimated_cost_reduction_per_hour_volume_usd REAL, -- Account-wide estimated volume-based USD/hour savings
    estimated_volume_reduction_per_hour REAL, -- Account-wide estimated volume reduction
    health TEXT, -- Health: DISABLED > INACTIVE > ERROR > OK
    inactive_services INTEGER, -- Services with INACTIVE health
    log_active_services INTEGER, -- Services not DISABLED or INACTIVE
    log_event_analyzed_count INTEGER, -- Number of log events that have been analyzed
    log_event_bytes_per_hour REAL, -- Discovered log event throughput in bytes/hour across all services
    log_event_cost_per_hour_bytes_usd REAL, -- Discovered log event ingestion cost in USD/hour across all services
    log_event_cost_per_hour_usd REAL, -- Discovered log event total cost in USD/hour across all services
    log_event_cost_per_hour_volume_usd REAL, -- Discovered log event indexing cost in USD/hour across all services
    log_event_count INTEGER, -- Total log events across all services
    log_event_quarantined_count INTEGER, -- Log events quarantined due to repeated AI analysis failures
    log_event_volume_per_hour REAL, -- Discovered log event throughput in events/hour across all services
    log_service_count INTEGER, -- Total number of services
    observed_bytes_per_hour_after REAL, -- Account-wide observed current bytes
    observed_bytes_per_hour_before REAL, -- Account-wide observed pre-approval bytes
    observed_cost_per_hour_after_bytes_usd REAL, -- Account-wide measured bytes-based USD/hour cost after approval
    observed_cost_per_hour_after_usd REAL, -- Account-wide measured total USD/hour cost after approval
    observed_cost_per_hour_after_volume_usd REAL, -- Account-wide measured volume-based USD/hour cost after approval
    observed_cost_per_hour_before_bytes_usd REAL, -- Account-wide measured bytes-based USD/hour cost before approval
    observed_cost_per_hour_before_usd REAL, -- Account-wide measured total USD/hour cost before approval
    observed_cost_per_hour_before_volume_usd REAL, -- Account-wide measured volume-based USD/hour cost before approval
    observed_volume_per_hour_after REAL, -- Account-wide observed current volume
    observed_volume_per_hour_before REAL, -- Account-wide observed pre-approval volume
    ok_services INTEGER, -- Services with OK health
    policy_approved_count INTEGER, -- Policies approved by user
    policy_dismissed_count INTEGER, -- Policies dismissed by user
    policy_pending_count INTEGER, -- Policies awaiting user action
    ready_for_use INTEGER, -- True when at least 1 log event has been analyzed
    refreshed_at TEXT,
    service_cost_per_hour_volume_usd REAL, -- Service-level indexing cost in USD/hour across all services
    service_volume_per_hour REAL, -- Ground-truth throughput in events/hour from service_log_volumes across all services
    warning TEXT, -- Most recent warning message
    warning_at TEXT -- When the most recent warning occurred
);

-- Datadog integration configuration for an account, one per account
CREATE TABLE datadog_accounts (
    id TEXT, -- Unique identifier of the Datadog configuration
    account_id TEXT, -- Parent account this configuration belongs to
    cost_per_gb_ingested REAL, -- Cost per GB of log data ingested (USD). NULL = using Datadog's published rate ($0.10/GB). Set to override with actual contract rate.
    created_at TEXT, -- When the Datadog account was created
    name TEXT, -- Display name for this Datadog account
    site TEXT, -- Datadog regional site. US1: datadoghq.com, US3: us3.datadoghq.com, US5: us5.datadoghq.com, EU1: datadoghq.eu, US1_FED: ddog-gov.com, AP1: ap1.datadoghq.com, AP2: ap2.datadoghq.com.
    updated_at TEXT -- When the Datadog account was last updated
);

-- Discovered Datadog log index where logs are stored (e.g., main, security, compliance)
CREATE TABLE datadog_log_indexes (
    id TEXT, -- Unique identifier for this index record
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from datadog_account.account_id.
    cost_per_million_events_indexed REAL, -- Cost per million events indexed in this index (USD). NULL = using Datadog's published rate ($1.70/M). SIEM indexes cost more — set accordingly.
    created_at TEXT, -- When this index was first discovered
    datadog_account_id TEXT, -- The Datadog account this index belongs to
    last_seen_at TEXT, -- Last time we saw logs flowing to this index
    name TEXT -- Index name from Datadog (e.g., 'main', 'security', 'compliance') - this is the stable identifier
);

-- Tracks health and progress of discovery operations from integration sources
CREATE TABLE discovery_statuses (
    id TEXT, -- Unique identifier of the discovery status record
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from datadog_account.account_id.
    completed_at TEXT, -- When the most recent iteration completed (successfully or with error)
    consecutive_errors INTEGER, -- Number of consecutive errors (reset on success)
    consecutive_warnings INTEGER, -- Number of consecutive warnings (reset on success)
    created_at TEXT, -- When status tracking began
    datadog_account_id TEXT, -- Datadog account performing the discovery (FK arc with other integrations)
    discovery_type TEXT, -- Type of discovery operation. service: discovers services in integration, log_events: discovers event patterns for a service, log_volume: calculates per-event volume, service_log_volume: calculates per-service volume over time.
    last_error TEXT, -- Last error message if discovery failed
    last_error_at TEXT, -- When the last error occurred
    last_warning TEXT, -- Last warning message (transient issues like rate limits)
    last_warning_at TEXT, -- When the last warning occurred
    service_id TEXT, -- Service being discovered (null for account-level service discovery)
    started_at TEXT, -- When the most recent iteration started
    updated_at TEXT -- When the status was last updated
);

-- Ground truth record for a field in a log event. Accumulates metadata as more production data is observed.
CREATE TABLE log_event_fields (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from log_event.account_id.
    created_at TEXT, -- When this field was first discovered
    distribution_observed_at TEXT, -- When value_distribution was last refreshed from production data.
    field_path TEXT, -- Unambiguous path segments, e.g. {attributes, http, status}
    log_event_id TEXT, -- The log event this field belongs to
    -- Top-N observed values with proportions. Populated on-demand for fields that need faceting (e.g., user agents for bot detection).
    -- Opaque JSON data. Query using SQLite json_extract() or json_each().
    value_distribution TEXT
);

-- AI-generated recommendation for a specific quality category on a log event, scoped to a workspace for approval
CREATE TABLE log_event_policies (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from workspace.account_id.
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
    -- $.low_value                    - Low value analysis (optional)
    -- $.accidental_debug_statements  - Accidental debug statements analysis (optional)
    -- $.malformed_data               - Malformed data analysis (optional)
    -- $.noise                        - Noise analysis (optional)
    -- $.noise.min_interval_seconds   number     - Suggested minimum interval between kept events in seconds
    -- $.duplicate_fields             - Duplicate fields analysis (optional)
    -- $.duplicate_fields.pairs[]     - List of duplicate field pairs
    -- $.duplicate_fields.pairs[].remove[] string[][] - List of duplicate field paths to remove
    -- $.duplicate_fields.pairs[].keep[] string[]   - Canonical field path to keep
    -- $.instrumentation_bloat        - Instrumentation bloat analysis (optional)
    -- $.instrumentation_bloat.fields[] string[][] - List of field paths that are instrumentation bloat
    -- $.oversized_fields             - Oversized fields analysis (optional)
    -- $.oversized_fields.fields[]    string[][] - List of field paths that are oversized
    -- 
    -- Example: json_extract(analysis, '$.field_name')
    analysis TEXT,
    approved_at TEXT, -- When this policy was approved by a user
    approved_by TEXT, -- User ID who approved this policy
    category TEXT, -- Quality issue category this policy addresses Values: health_checks, bot_traffic, low_value, accidental_debug_statements, malformed_data, noise, pii_leakage, secrets_leakage, phi_leakage, payment_data_leakage, duplicate_fields, instrumentation_bloat, oversized_fields.
    category_type TEXT, -- Type of problem: compliance (legal/security risk) or waste (cost reduction). Auto-set via trigger from CategoryMeta.
    created_at TEXT, -- When this policy was created
    dismissed_at TEXT, -- When this policy was dismissed by a user
    dismissed_by TEXT, -- User ID who dismissed this policy
    log_event_id TEXT, -- The log event this policy applies to
    model TEXT, -- AI model that generated this policy (e.g., 'claude-sonnet-4-20250514')
    subjective INTEGER, -- Whether this category requires AI judgment (true) vs mechanically verifiable (false). Auto-set via trigger from CategoryMeta.
    updated_at TEXT, -- When this policy was last updated
    workspace_id TEXT -- The workspace that owns this policy
);

-- Cache table for log_event_policy_statuses view. Refreshed by cron service.
CREATE TABLE log_event_policy_statuses_cache (
    id TEXT,
    account_id TEXT,
    approved_at TEXT,
    category TEXT,
    category_type TEXT, -- Values: compliance, waste.
    created_at TEXT,
    dismissed_at TEXT,
    estimated_bytes_reduction_per_hour REAL, -- Bytes/hour saved if this policy applied alone. NULL if not estimable.
    estimated_cost_reduction_per_hour_bytes_usd REAL,
    estimated_cost_reduction_per_hour_usd REAL,
    estimated_cost_reduction_per_hour_volume_usd REAL,
    estimated_volume_reduction_per_hour REAL, -- Events/hour saved if this policy applied alone. NULL if not estimable.
    impact_type TEXT, -- How policy achieves cost reduction: attribute (bytes only), volume (events), none (compliance only)
    log_event_id TEXT,
    policy_id TEXT,
    refreshed_at TEXT,
    status TEXT, -- Values: PENDING, APPROVED, DISMISSED.
    subjective INTEGER,
    survival_rate REAL, -- Fraction of events that survive this policy (0.0 = all dropped, 1.0 = all kept). NULL if not estimable.
    workspace_id TEXT
);

-- Cache table for log_event_statuses view. Refreshed by cron service.
CREATE TABLE log_event_statuses_cache (
    id TEXT,
    account_id TEXT, -- Account ID for tenant isolation
    approved_policy_count INTEGER, -- Policies approved by user
    bytes_per_hour REAL, -- Current throughput in bytes/hour (rolling 7-day)
    cost_per_hour_bytes_usd REAL, -- Current ingestion cost in USD/hour
    cost_per_hour_usd REAL, -- Current total cost in USD/hour (bytes + volume)
    cost_per_hour_volume_usd REAL, -- Current indexing cost in USD/hour
    dismissed_policy_count INTEGER, -- Policies dismissed by user
    error TEXT, -- Error message when is_broken = true
    estimated_bytes_reduction_per_hour REAL, -- Bytes/hour saved by all policies combined
    estimated_cost_reduction_per_hour_bytes_usd REAL, -- Estimated ingestion savings in USD/hour
    estimated_cost_reduction_per_hour_usd REAL, -- Estimated total savings in USD/hour (bytes + volume)
    estimated_cost_reduction_per_hour_volume_usd REAL, -- Estimated indexing savings in USD/hour
    estimated_volume_reduction_per_hour REAL, -- Events/hour saved by all policies combined
    has_been_analyzed INTEGER, -- Whether AI has analyzed this log event
    has_volumes INTEGER, -- Whether volume data exists for this log event
    is_broken INTEGER, -- Whether discovery has consecutive errors for this service
    is_quarantined INTEGER, -- Whether this log event is excluded from analysis due to repeated failures
    log_event_id TEXT, -- The log event this status belongs to
    observed_bytes_per_hour_after REAL, -- Measured bytes/hour after policy approval (current)
    observed_bytes_per_hour_before REAL, -- Measured bytes/hour before first policy approval
    observed_cost_per_hour_after_bytes_usd REAL, -- Measured ingestion cost after approval (current)
    observed_cost_per_hour_after_usd REAL, -- Measured total cost after approval (current)
    observed_cost_per_hour_after_volume_usd REAL, -- Measured indexing cost after approval (current)
    observed_cost_per_hour_before_bytes_usd REAL, -- Measured ingestion cost before approval
    observed_cost_per_hour_before_usd REAL, -- Measured total cost before approval
    observed_cost_per_hour_before_volume_usd REAL, -- Measured indexing cost before approval
    observed_volume_per_hour_after REAL, -- Measured events/hour after policy approval (current)
    observed_volume_per_hour_before REAL, -- Measured events/hour before first policy approval
    pending_policy_count INTEGER, -- Policies awaiting user action
    policy_count INTEGER, -- Total non-dismissed policies
    refreshed_at TEXT,
    service_id TEXT, -- Service ID (denormalized from log_event)
    volume_per_hour REAL -- Current throughput in events/hour (rolling 7-day)
);

-- Distinct log message pattern discovered within a service. Defines how to parse and match logs using codecs and matchers.
CREATE TABLE log_events (
    id TEXT, -- Unique identifier of the log event
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from service.account_id.
    analyzed_at TEXT, -- When AI last analyzed this log event for quality issues
    created_at TEXT, -- When the log event was created
    description TEXT, -- What this event pattern represents
    -- Sample log records captured during discovery, used for AI analysis and pattern validation
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
    -- $[0].match_type                string     - Match operator: exact, contains, starts_with, ends_with, regex, exists
    -- $[0].match_value               string     - Value to match against
    -- $[0].case_insensitive          boolean    - Whether matching is case-insensitive (optional)
    -- $[0].negate                    boolean    - Whether to invert match result (optional)
    -- 
    -- Example: json_extract(matchers, '$[0].field_name')
    matchers TEXT,
    name TEXT, -- Snake_case identifier unique per service, e.g. nginx_access_log
    service_id TEXT, -- Service that produces this event
    updated_at TEXT -- When the log event was last updated
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

-- Cache table for service_statuses view. Refreshed by cron service.
CREATE TABLE service_statuses_cache (
    id TEXT,
    account_id TEXT, -- Account ID (denormalized from service)
    datadog_account_id TEXT, -- The Datadog account performing discovery
    error TEXT, -- Most recent error from discovery (when health = ERROR)
    error_at TEXT, -- When the error occurred
    estimated_bytes_reduction_per_hour REAL, -- Estimated bytes reduction from active policies
    estimated_cost_reduction_per_hour_bytes_usd REAL, -- Estimated bytes-based USD/hour savings from active policies
    estimated_cost_reduction_per_hour_usd REAL, -- Estimated total USD/hour savings from active policies
    estimated_cost_reduction_per_hour_volume_usd REAL, -- Estimated volume-based USD/hour savings from active policies
    estimated_volume_reduction_per_hour REAL, -- Estimated volume reduction from active policies
    health TEXT, -- Health: DISABLED > INACTIVE > ERROR > OK
    log_event_analyzed_count INTEGER, -- Number of log events that have been analyzed
    log_event_bytes_per_hour REAL, -- Discovered log event throughput in bytes/hour from rolling 7-day window
    log_event_cost_per_hour_bytes_usd REAL, -- Discovered log event ingestion cost in USD/hour
    log_event_cost_per_hour_usd REAL, -- Discovered log event total cost in USD/hour (bytes + volume)
    log_event_cost_per_hour_volume_usd REAL, -- Discovered log event indexing cost in USD/hour
    log_event_count INTEGER, -- Total number of log events discovered for this service
    log_event_quarantined_count INTEGER, -- Log events quarantined due to repeated AI analysis failures
    log_event_volume_per_hour REAL, -- Discovered log event throughput in events/hour from rolling 7-day window
    observed_bytes_per_hour_after REAL, -- Measured bytes/hour after policy approval
    observed_bytes_per_hour_before REAL, -- Measured bytes/hour before first policy approval
    observed_cost_per_hour_after_bytes_usd REAL, -- Measured bytes-based USD/hour cost after approval
    observed_cost_per_hour_after_usd REAL, -- Measured total USD/hour cost after approval
    observed_cost_per_hour_after_volume_usd REAL, -- Measured volume-based USD/hour cost after approval
    observed_cost_per_hour_before_bytes_usd REAL, -- Measured bytes-based USD/hour cost before approval
    observed_cost_per_hour_before_usd REAL, -- Measured total USD/hour cost before approval
    observed_cost_per_hour_before_volume_usd REAL, -- Measured volume-based USD/hour cost before approval
    observed_volume_per_hour_after REAL, -- Measured events/hour after policy approval
    observed_volume_per_hour_before REAL, -- Measured events/hour before first policy approval
    policy_approved_count INTEGER, -- Policies approved by user
    policy_dismissed_count INTEGER, -- Policies dismissed by user
    policy_pending_count INTEGER, -- Policies awaiting user action
    refreshed_at TEXT,
    service_cost_per_hour_volume_usd REAL, -- Service-level indexing cost in USD/hour based on total service volume
    service_id TEXT, -- The service this status belongs to
    service_volume_per_hour REAL, -- Ground-truth service throughput in events/hour from service_log_volumes rolling 7-day window
    warning TEXT, -- Most recent warning (rate limits, etc.)
    warning_at TEXT -- When the warning occurred
);

-- Application or microservice that produces logs. Central entity in the data catalog.
CREATE TABLE services (
    id TEXT, -- Unique identifier of the service
    account_id TEXT, -- Parent account this service belongs to
    created_at TEXT, -- When the service was created
    description TEXT, -- AI-generated description of what this service does and its telemetry characteristics
    enabled INTEGER, -- Whether log analysis and policy generation is active for this service
    initial_weekly_log_count INTEGER, -- Approximate weekly log count from initial discovery (7-day period from Datadog)
    name TEXT, -- Service identifier in telemetry (e.g., 'checkout-service')
    updated_at TEXT -- When the service was last updated
);

-- Group of users within a workspace that reviews policies and manages services
CREATE TABLE teams (
    id TEXT, -- Unique identifier of the team
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from workspace.account_id.
    created_at TEXT, -- When the team was created
    name TEXT, -- Human-readable name within the workspace
    updated_at TEXT, -- When the team was last updated
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
    purpose TEXT, -- Primary purpose determining evaluation strategy. observability: performance and reliability, security: threat detection, compliance: regulatory requirements.
    updated_at TEXT -- When the workspace was last updated
);

