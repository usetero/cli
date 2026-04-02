package powersync

// clientIndexes defines indexes for client-side query performance.
// These are not part of the server schema — they optimize local SQLite queries.
// The map key is the table name, the value is a list of indexes to create.
var clientIndexes = map[string][]SchemaIndex{
	"services": {
		{Name: "name", Columns: []SchemaIndexColumn{
			{Name: "name", Ascending: true, Type: "text"},
		}},
	},

	"log_events": {
		{Name: "service_id", Columns: []SchemaIndexColumn{
			{Name: "service_id", Ascending: true, Type: "text"},
		}},
	},

	// Filtered by status, category, objectivity, risk_level in AI-generated queries.
	// Joined on log_event_id.
	"log_event_policy_statuses_cache": {
		{Name: "log_event_id", Columns: []SchemaIndexColumn{
			{Name: "log_event_id", Ascending: true, Type: "text"},
		}},
		{Name: "category_status", Columns: []SchemaIndexColumn{
			{Name: "category", Ascending: true, Type: "text"},
			{Name: "status", Ascending: true, Type: "text"},
		}},
	},

	"log_event_statuses_cache": {
		{Name: "log_event_id", Columns: []SchemaIndexColumn{
			{Name: "log_event_id", Ascending: true, Type: "text"},
		}},
		{Name: "service_id", Columns: []SchemaIndexColumn{
			{Name: "service_id", Ascending: true, Type: "text"},
		}},
	},

	"service_statuses_cache": {
		{Name: "service_id", Columns: []SchemaIndexColumn{
			{Name: "service_id", Ascending: true, Type: "text"},
		}},
	},

	"log_event_policies": {
		{Name: "log_event_id", Columns: []SchemaIndexColumn{
			{Name: "log_event_id", Ascending: true, Type: "text"},
		}},
	},

	"messages": {
		{Name: "conversation_id", Columns: []SchemaIndexColumn{
			{Name: "conversation_id", Ascending: true, Type: "text"},
		}},
	},

	"conversations": {
		{Name: "account_id", Columns: []SchemaIndexColumn{
			{Name: "account_id", Ascending: true, Type: "text"},
		}},
	},

	"datadog_account_statuses_cache": {
		{Name: "datadog_account_id", Columns: []SchemaIndexColumn{
			{Name: "datadog_account_id", Ascending: true, Type: "text"},
		}},
	},
}

// ApplyClientIndexes merges client-side indexes into schema tables.
// Tables without indexes get an empty slice so JSON encodes as [] not null.
func ApplyClientIndexes(tables []SchemaTable) []SchemaTable {
	for i, table := range tables {
		if indexes, ok := clientIndexes[table.Name]; ok {
			tables[i].Indexes = indexes
		} else {
			tables[i].Indexes = []SchemaIndex{}
		}
	}
	return tables
}

func applyClientIndexes(tables []SchemaTable) []SchemaTable {
	return ApplyClientIndexes(tables)
}
