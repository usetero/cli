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

// knownTables maps PowerSync internal names to public Table constants.
var knownTables = map[string]Table{
	"messages":          TableMessages,
	"ps_data__messages": TableMessages,

	"conversations":          TableConversations,
	"ps_data__conversations": TableConversations,
}
