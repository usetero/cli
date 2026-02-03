package tool

type QueryInput struct {
	SQL string `json:"sql"`
}

type QueryResult struct {
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
}
