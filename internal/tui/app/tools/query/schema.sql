-- Tero Local Catalog Schema
-- Auto-generated with comments from control plane. DO NOT EDIT.
--
-- SQLite 3.51.1 with full support for: CTEs, window functions, JSON functions,
-- RETURNING, generated columns, and all modern SQL features.
--
-- This is a LOCAL SQLite database synced from the Tero control plane.
-- All data is scoped to the authenticated user's account.
--
-- Key concepts:
--   services      - Applications producing logs (e.g., 'checkout-service')
--   log_events    - Distinct event patterns within a service
--   policies      - AI-identified waste (health checks, duplicate fields, bloat)
--   *_cache       - Pre-computed status and metrics (query these for current state)
--
-- IMPORTANT: All queries are READ-ONLY. This is a local sync of server data.


CREATE TABLE conversation_contexts (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from conversation.account_id.
    added_by TEXT, -- Who added this entity to context
    conversation_id TEXT, -- Conversation this context belongs to
    created_at TEXT, -- When the entity was added to context
    entity_id TEXT, -- ID of the context entity
    entity_type TEXT -- Type of the context entity
);

CREATE TABLE conversations (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from workspace.account_id.
    created_at TEXT, -- When the conversation was created
    title TEXT, -- AI-generated title, set after first exchange
    updated_at TEXT, -- When the conversation was last updated
    user_id TEXT, -- WorkOS user ID who owns this conversation
    workspace_id TEXT -- Workspace this conversation belongs to
);

CREATE TABLE datadog_account_statuses_cache (
    id TEXT,
    account_id TEXT, -- Account ID (denormalized from Datadog account)
    datadog_account_id TEXT, -- The Datadog account this status belongs to
    log_active_services INTEGER, -- Services not DISABLED or INACTIVE
    log_analyzed_count INTEGER, -- Number of log events that have been analyzed (saved + valuable + waste)
    log_analyzing_count INTEGER, -- Log events with ANALYZING status
    log_analyzing_services INTEGER, -- Services with ANALYZING status
    log_broken_services INTEGER, -- Services with BROKEN status
    log_disabled_services INTEGER, -- Services with DISABLED status
    log_discovered_volume TEXT,
    log_discovering_count INTEGER, -- Log events with DISCOVERING status
    log_discovering_services INTEGER, -- Services with DISCOVERING status
    log_error TEXT, -- Most recent error message from any broken service or account-level discovery
    log_error_at TEXT, -- When the most recent error occurred
    log_event_count INTEGER, -- Total log events across all services
    log_inactive_services INTEGER, -- Services with INACTIVE status
    log_percent_complete REAL, -- Overall coverage percentage (discovered / service volume * 100)
    log_ready_services INTEGER, -- Services with READY status
    log_saved_count INTEGER, -- Log events with SAVED status (policy approved)
    log_service_count INTEGER, -- Total number of services
    log_service_volume TEXT,
    log_stale_services INTEGER, -- Services with STALE status
    log_status TEXT, -- Status: DISABLED > INACTIVE > BROKEN > STALE > DISCOVERING > ANALYZING > READY
    log_valuable_count INTEGER, -- Log events with VALUABLE status (analyzed, no issues)
    log_warning TEXT, -- Most recent warning message (e.g., rate limit)
    log_warning_at TEXT, -- When the most recent warning occurred
    log_waste_count INTEGER, -- Log events with WASTE status (issues found, pending action)
    ready_for_use INTEGER, -- True when account has enough analyzed data for good UX (>= 50 analyzed log events)
    refreshed_at TEXT
);

CREATE TABLE datadog_accounts (
    id TEXT, -- Unique identifier of the Datadog configuration
    account_id TEXT, -- Parent account this configuration belongs to
    created_at TEXT, -- When the Datadog account was created
    name TEXT, -- Display name for this Datadog account
    site TEXT, -- Datadog site for this account (must be explicit)
    updated_at TEXT -- When the Datadog account was last updated
);

CREATE TABLE datadog_log_indexes (
    id TEXT, -- Unique identifier for this index record
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from datadog_account.account_id.
    created_at TEXT, -- When this index was first discovered
    datadog_account_id TEXT, -- The Datadog account this index belongs to
    last_seen_at TEXT, -- Last time we saw logs flowing to this index
    name TEXT -- Index name from Datadog (e.g., 'main', 'security', 'compliance') - this is the stable identifier
);

CREATE TABLE discovery_statuses (
    id TEXT, -- Unique identifier of the discovery status record
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from datadog_account.account_id.
    completed_at TEXT, -- When the most recent discovery run completed (successfully or with error)
    consecutive_errors INTEGER, -- Number of consecutive errors (reset on success)
    consecutive_warnings INTEGER, -- Number of consecutive warnings (reset on success)
    created_at TEXT, -- When status tracking began
    datadog_account_id TEXT, -- Datadog account performing the discovery (FK arc with other integrations)
    discovery_type TEXT, -- Type of discovery operation being tracked
    last_error TEXT, -- Last error message if discovery failed
    last_error_at TEXT, -- When the last error occurred
    last_warning TEXT, -- Last warning message (transient issues like rate limits)
    last_warning_at TEXT, -- When the last warning occurred
    service_id TEXT, -- Service being discovered (null for account-level service discovery)
    started_at TEXT, -- When the current/most recent discovery run started
    updated_at TEXT -- When the status was last updated
);

CREATE TABLE log_event_policies (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from workspace.account_id.
    analysis TEXT, -- Category-specific analysis (only one field populated, matching category)
    approved_at TEXT, -- When this policy was approved by a user
    approved_by TEXT, -- User ID who approved this policy
    category TEXT, -- Policy category (e.g., 'health_checks', 'pii_leakage')
    created_at TEXT, -- When this policy was created
    dismissed_at TEXT, -- When this policy was dismissed by a user
    dismissed_by TEXT, -- User ID who dismissed this policy
    log_event_id TEXT, -- The log event this policy applies to
    model TEXT, -- AI model that generated this policy (e.g., 'claude-sonnet-4-20250514')
    updated_at TEXT, -- When this policy was last updated
    workspace_id TEXT -- The workspace that owns this policy
);

CREATE TABLE log_event_policy_statuses_cache (
    id TEXT,
    account_id TEXT, -- Account ID for tenant isolation
    approved_at TEXT, -- When the policy was approved (anchor for savings calculation)
    bytes_saved TEXT,
    category TEXT, -- Policy category (e.g., health_checks, duplicate_fields, instrumentation_bloat)
    dismissed_at TEXT, -- When the policy was dismissed
    log_event_id TEXT, -- The log event this policy applies to
    policy_id TEXT, -- The policy this status belongs to
    refreshed_at TEXT,
    status TEXT, -- WASTE: found waste waiting for action, SAVED: user approved, DISMISSED: user rejected
    volume_saved TEXT,
    workspace_id TEXT -- The workspace that owns this policy
);

CREATE TABLE log_event_statuses_cache (
    id TEXT,
    account_id TEXT, -- Account ID for tenant isolation
    bytes_after TEXT,
    bytes_before TEXT,
    datadog_account_id TEXT, -- The Datadog account performing discovery
    error TEXT, -- Error message if status is BROKEN (5+ consecutive discovery failures)
    has_been_analyzed INTEGER, -- Whether this log event has been analyzed for quality issues
    has_volumes INTEGER, -- Whether log_event_volumes exist for this event from this integration
    log_event_id TEXT, -- The log event this status belongs to
    refreshed_at TEXT,
    service_id TEXT, -- Service ID (denormalized from log_event)
    status TEXT, -- Unified status: BROKEN > SAVED > VALUABLE > WASTE > ANALYZING > DISCOVERING
    volume_after TEXT,
    volume_before TEXT
);

CREATE TABLE log_event_volumes (
    id TEXT, -- Unique identifier of this volume record
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from log_event.account_id.
    attribute_avg_bytes TEXT, -- Average bytes per attribute name, estimated from sampled logs
    avg_bytes REAL, -- Average bytes per log in this hour bucket, estimated from sampled logs
    count TEXT,
    created_at TEXT, -- When this volume record was created
    datadog_log_index_id TEXT, -- The Datadog log index where this volume was observed
    edge_instance_id TEXT, -- The edge instance where this volume was observed
    log_event_id TEXT, -- The log event this volume measurement belongs to
    timestamp TEXT -- Hour bucket timestamp (truncated to beginning of hour)
);

CREATE TABLE log_events (
    id TEXT, -- Unique identifier of the log event
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from service.account_id.
    analyzed_at TEXT, -- When AI last analyzed this log event for quality issues
    created_at TEXT, -- When the log event was created
    description TEXT, -- What this event pattern represents
    examples TEXT, -- Example logs with timestamp, message, and attributes
    matchers TEXT, -- Structured matchers to identify this event
    name TEXT, -- Snake_case identifier for event type
    service_id TEXT, -- Service that produces this event
    updated_at TEXT -- When the log event was last updated
);

CREATE TABLE messages (
    id TEXT, -- Unique identifier
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from conversation.account_id.
    content TEXT, -- Array of typed content blocks: text, thinking, tool_use, tool_result
    conversation_id TEXT, -- Conversation this message belongs to
    created_at TEXT, -- When the message was created
    model TEXT, -- AI model that produced this message. Null for user messages.
    role TEXT, -- Who sent this message.
    stop_reason TEXT -- Why the assistant stopped: end_turn, tool_use. Null for user messages.
);

CREATE TABLE service_log_volumes (
    id TEXT, -- Unique identifier of this volume record
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from service.account_id.
    created_at TEXT, -- When this volume record was created
    datadog_account_id TEXT, -- Datadog account that reported this service volume
    service_id TEXT, -- The service this volume tracking belongs to
    timestamp TEXT, -- Hour boundary for this volume bucket (truncated to hour)
    updated_at TEXT, -- When this volume record was last updated
    volume TEXT
);

CREATE TABLE service_statuses_cache (
    id TEXT,
    account_id TEXT, -- Account ID (denormalized from service)
    datadog_account_id TEXT, -- The Datadog account performing discovery
    log_analyzing_count INTEGER, -- Number of log events with ANALYZING status
    log_bytes_after TEXT,
    log_bytes_before TEXT,
    log_discovered_volume TEXT,
    log_discovering_count INTEGER, -- Number of log events with DISCOVERING status
    log_error TEXT, -- Error message if log_status is BROKEN
    log_error_at TEXT, -- When the error occurred
    log_event_count INTEGER, -- Total number of log events discovered for this service
    log_percent_complete REAL, -- Coverage percentage (discovered_volume / service_volume * 100)
    log_saved_count INTEGER, -- Number of log events with SAVED status (policy approved)
    log_service_volume TEXT,
    log_status TEXT, -- Status: DISABLED > INACTIVE > BROKEN > STALE > DISCOVERING > ANALYZING > READY
    log_valuable_count INTEGER, -- Number of log events with VALUABLE status (analyzed, no issues)
    log_volume_after TEXT,
    log_volume_before TEXT,
    log_warning TEXT, -- Warning message (e.g., rate limit) - informational, does not affect status
    log_warning_at TEXT, -- When the warning occurred
    log_waste_count INTEGER, -- Number of log events with WASTE status (issues found, pending action)
    refreshed_at TEXT,
    service_id TEXT -- The service this status belongs to
);

CREATE TABLE services (
    id TEXT, -- Unique identifier of the service
    account_id TEXT, -- Parent account this service belongs to
    created_at TEXT, -- When the service was created
    description TEXT, -- AI-generated description of what this service does and its telemetry characteristics
    enabled INTEGER, -- Whether telemetry analysis is enabled
    initial_weekly_log_count INTEGER, -- Approximate weekly log count from initial discovery (7-day period from Datadog)
    name TEXT, -- Service identifier in telemetry (e.g., 'checkout-service')
    updated_at TEXT -- When the service was last updated
);

CREATE TABLE teams (
    id TEXT, -- Unique identifier of the team
    account_id TEXT, -- Denormalized for tenant isolation. Auto-set via trigger from workspace.account_id.
    created_at TEXT, -- When the team was created
    name TEXT, -- Human-readable name within the workspace
    updated_at TEXT, -- When the team was last updated
    workspace_id TEXT -- Parent workspace this team belongs to
);

CREATE TABLE workspaces (
    id TEXT, -- Unique identifier of the workspace
    account_id TEXT, -- Parent account this workspace belongs to
    created_at TEXT, -- When the workspace was created
    name TEXT, -- Human-readable name within the account
    purpose TEXT, -- Primary purpose determining evaluation strategy
    updated_at TEXT -- When the workspace was last updated
);

