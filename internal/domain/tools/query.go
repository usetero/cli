package tools

// QueryInput is the input schema for the query tool.
type QueryInput struct {
	SQL string `json:"sql"`
}

// QueryResult is the typed output of a query tool execution.
type QueryResult struct {
	Rows []map[string]any `json:"rows"`
}

// ToMap serializes the result for the GraphQL API.
func (r QueryResult) ToMap() map[string]any {
	return map[string]any{"rows": r.Rows}
}
