package tool

type ShowTableInput struct {
	Title   string   `json:"title"`
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
}
