package block

import "fmt"

// QueryInput runs a SQL query against the catalog.
type QueryInput struct {
	SQL string `json:"sql"`
}

// Validate checks that required fields are present.
func (i QueryInput) Validate() error {
	if i.SQL == "" {
		return fmt.Errorf("sql is required")
	}
	return nil
}

// QueryResult is the result of a query.
type QueryResult struct {
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
}
