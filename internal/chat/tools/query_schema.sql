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
    health TEXT, -- Health: DISABLED > INACTIVE > ERROR > STALE > OK
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
    stale_services INTEGER, -- Services with STALE health
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
    avg_bytes REAL, -- Average serialized byte size of this field's value. Initially estimated from example log records, refined by trigger from log_event_volumes data.
    created_at TEXT, -- When this field was discovered
    distribution_observed_at TEXT, -- When value_distribution was last refreshed from production data.
    field_path TEXT, -- Unambiguous path segments, e.g. {attributes, http, status}
    log_event_id TEXT, -- The log event this field belongs to
    value_distribution TEXT -- Top-N observed values with proportions. Populated on-demand for fields that need faceting (e.g., user agents for bot detection).
);

-- AI-generated recommendation for a specific quality category on a log event, scoped to a workspace for approval
CREATE TABLE log_event_policies (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from workspace.account_id.
    analysis TEXT, -- Category-specific analysis from AI. JSON object with one field populated matching the category, containing the analysis and recommended actions.
    approved_at TEXT, -- When this policy was approved by a user
    approved_by TEXT, -- User ID who approved this policy
    benefits TEXT, -- What benefits this policy provides. volume_reduction: fewer events, bytes_reduction: smaller events, signal_quality: less noise, compliance: regulatory/policy, resilience: system stability.
    category TEXT, -- Quality issue category this policy addresses Values: health_checks, bot_traffic, low_value, accidental_debug_statements, malformed_data, noise, pii_leakage, duplicate_fields, instrumentation_bloat, oversized_fields.
    created_at TEXT, -- When this policy was created
    dismissed_at TEXT, -- When this policy was dismissed by a user
    dismissed_by TEXT, -- User ID who dismissed this policy
    log_event_id TEXT, -- The log event this policy applies to
    model TEXT, -- AI model that generated this policy (e.g., 'claude-sonnet-4-20250514')
    objectivity TEXT, -- How verifiable is this policy? factual: user can confirm by looking at the data, reasoned: requires AI judgment. Auto-set by trigger from category.
    updated_at TEXT, -- When this policy was last updated
    workspace_id TEXT -- The workspace that owns this policy
);

-- Cache table for log_event_policy_statuses view. Refreshed by cron service.
CREATE TABLE log_event_policy_statuses_cache (
    id TEXT,
    account_id TEXT, -- Account ID for tenant isolation
    approved_at TEXT, -- When the policy was approved
    benefits TEXT, -- What this policy provides (volume_reduction, bytes_reduction, signal_quality, compliance, resilience)
    category TEXT, -- Quality issue category (e.g., health_checks, pii_leakage, duplicate_fields)
    created_at TEXT, -- When this policy was created
    dismissed_at TEXT, -- When the policy was dismissed
    estimated_bytes_reduction_per_hour REAL, -- Projected bytes/hour eliminated. NULL when not estimable.
    estimated_cost_reduction_per_hour_bytes_usd REAL, -- Projected USD/hour savings from bytes reduction (ingestion cost component). NULL when not estimable.
    estimated_cost_reduction_per_hour_usd REAL, -- Projected USD/hour savings from this policy (bytes + volume combined). Aggregated across all datadog accounts. NULL when not estimable.
    estimated_cost_reduction_per_hour_volume_usd REAL, -- Projected USD/hour savings from volume reduction (indexing cost component). Zero for attribute-only policies. NULL when not estimable.
    estimated_volume_reduction_per_hour REAL, -- Projected events/hour eliminated. NULL when not estimable (compliance, resilience policies).
    log_event_id TEXT, -- The log event this policy applies to
    objectivity TEXT, -- How verifiable: factual (user can confirm) or reasoned (AI judgment)
    policy_id TEXT, -- The policy this status belongs to
    refreshed_at TEXT,
    status TEXT, -- PENDING: awaiting action, APPROVED: user approved, DISMISSED: user rejected
    workspace_id TEXT -- The workspace that owns this policy
);

-- Cache table for log_event_statuses view. Refreshed by cron service.
CREATE TABLE log_event_statuses_cache (
    id TEXT,
    account_id TEXT, -- Account ID for tenant isolation
    approved_policy_count INTEGER, -- Policies approved by user
    bytes_per_hour REAL, -- Current throughput in bytes/hour from the rolling 7-day window. Same baseline used for policy impact estimates, so directly comparable. NULL when no volume data exists.
    cost_per_hour_bytes_usd REAL, -- Current bytes-based cost in USD/hour (ingestion component). NULL when no volume data exists.
    cost_per_hour_usd REAL, -- Current cost in USD/hour (bytes + volume combined). Independent of policies. NULL when no volume data exists.
    cost_per_hour_volume_usd REAL, -- Current volume-based cost in USD/hour (indexing component). NULL when no volume data exists.
    datadog_account_id TEXT, -- The Datadog account performing discovery
    dismissed_policy_count INTEGER, -- Policies dismissed by user
    error TEXT, -- Error message if status is BROKEN
    estimable_policy_count INTEGER, -- Active policies whose category is estimable (deterministic categories where baseline = estimate). See is_estimable_category() in triggers/06_category_functions.sql.
    estimated_bytes_reduction_per_hour REAL, -- Sum of estimated bytes reductions from all active estimable policies
    estimated_cost_reduction_per_hour_bytes_usd REAL, -- Estimated USD/hour savings from bytes reduction (ingestion component).
    estimated_cost_reduction_per_hour_usd REAL, -- Estimated USD/hour savings from active policies (bytes + volume combined).
    estimated_cost_reduction_per_hour_volume_usd REAL, -- Estimated USD/hour savings from volume reduction (indexing component).
    estimated_volume_reduction_per_hour REAL, -- Sum of estimated volume reductions from all active estimable policies
    has_been_analyzed INTEGER, -- Whether this log event has been analyzed for quality issues
    has_volumes INTEGER, -- Whether log_event_volumes exist for this event from this integration
    log_event_id TEXT, -- The log event this status belongs to
    observed_bytes_per_hour_after REAL, -- Measured bytes/hour in recent 7-day window
    observed_bytes_per_hour_before REAL, -- Measured bytes/hour before first policy approval
    observed_cost_per_hour_after_bytes_usd REAL, -- Measured bytes-based USD/hour cost in recent window after policy approval. Only when approved and data is fresh.
    observed_cost_per_hour_after_usd REAL, -- Measured USD/hour cost in recent window after policy approval (bytes + volume combined). Only when approved and data is fresh.
    observed_cost_per_hour_after_volume_usd REAL, -- Measured volume-based USD/hour cost in recent window after policy approval. Only when approved and data is fresh.
    observed_cost_per_hour_before_bytes_usd REAL, -- Measured bytes-based USD/hour cost before first policy approval. Only when approved and data is fresh.
    observed_cost_per_hour_before_usd REAL, -- Measured USD/hour cost before first policy approval (bytes + volume combined). Only when approved and data is fresh.
    observed_cost_per_hour_before_volume_usd REAL, -- Measured volume-based USD/hour cost before first policy approval. Only when approved and data is fresh.
    observed_volume_per_hour_after REAL, -- Measured events/hour in recent 7-day window (only when approved and data is fresh)
    observed_volume_per_hour_before REAL, -- Measured events/hour in 7-day window before first policy approval (only when approved and data is fresh)
    pending_policy_count INTEGER, -- Policies awaiting user action
    policy_count INTEGER, -- Total active (non-dismissed) policies on this log event
    refreshed_at TEXT,
    service_id TEXT, -- Service ID (denormalized from log_event)
    status TEXT, -- BROKEN: discovery errors, QUARANTINED: repeated AI analysis failures, RESOLVED: all policies acted on, CLEAN: analyzed with no issues, PENDING: has policies awaiting action, ANALYZING: AI working, DISCOVERING: collecting data
    volume_per_hour REAL -- Current throughput in events/hour from the rolling 7-day window. Same baseline used for policy impact estimates, so directly comparable. NULL when no volume data exists.
);

-- Distinct log message pattern discovered within a service. Defines how to parse and match logs using codecs and matchers.
CREATE TABLE log_events (
    id TEXT, -- Unique identifier of the log event
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from service.account_id.
    analyzed_at TEXT, -- When AI last analyzed this log event for quality issues
    created_at TEXT, -- When the log event was created
    description TEXT, -- What this event pattern represents
    examples TEXT, -- Sample log records captured during discovery, used for AI analysis and pattern validation
    matchers TEXT, -- JSON rules that match incoming logs to this event. Each matcher specifies a field path, operator, and value.
    name TEXT, -- Snake_case identifier unique per service, e.g. nginx_access_log
    service_id TEXT, -- Service that produces this event
    updated_at TEXT -- When the log event was last updated
);

-- Single message in a chat conversation. Append-only — never updated or deleted.
CREATE TABLE messages (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from conversation.account_id.
    content TEXT, -- Array of typed content blocks: text, thinking, tool_use, tool_result
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
    health TEXT, -- Health: DISABLED > INACTIVE > ERROR > STALE > OK
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

