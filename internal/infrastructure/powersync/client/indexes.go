package client

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

func applyClientIndexes(tables []SchemaTable) []SchemaTable {
	for i, table := range tables {
		if indexes, ok := clientIndexes[table.Name]; ok {
			tables[i].Indexes = indexes
		} else {
			tables[i].Indexes = []SchemaIndex{}
		}
	}
	return tables
}
