package sqlite

// Table represents a watched table.
type Table string

// Table names matching the schema.
const (
	TableConversations Table = "conversations"
	TableMessages      Table = "messages"
)

// knownTables maps PowerSync internal names to public Table constants.
var knownTables = map[string]Table{
	"messages":          TableMessages,
	"ps_data__messages": TableMessages,

	"conversations":          TableConversations,
	"ps_data__conversations": TableConversations,
}
