package tools

// QueryInput is the input schema for the query tool.
type QueryInput struct {
	SQL    string `json:"sql"`
	Status string `json:"status"`
	Result string `json:"result"`
}

// QueryResult is the typed output of a query tool execution.
type QueryResult struct {
	Rows        []map[string]any `json:"rows"`
	RowsDropped int              `json:"rows_dropped,omitempty"`
}

// ToMap serializes the result for the GraphQL API.
func (r QueryResult) ToMap() map[string]any {
	m := map[string]any{"rows": r.Rows}
	if r.RowsDropped > 0 {
		m["rows_dropped"] = r.RowsDropped
	}
	return m
}
