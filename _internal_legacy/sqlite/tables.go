package sqlite

// Table represents a watched table.
type Table string

// Table names matching the schema.
const (
	TableConversations    Table = "conversations"
	TableLogEventPolicies Table = "log_event_policies"
	TableLogEvents        Table = "log_events"
	TableMessages         Table = "messages"
	TableServices         Table = "services"
)
