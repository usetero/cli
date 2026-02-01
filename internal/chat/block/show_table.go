package block

import "fmt"

// ShowTableInput displays a table of data.
type ShowTableInput struct {
	Title   string   `json:"title"`
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
}

// Validate checks that required fields are present.
func (i ShowTableInput) Validate() error {
	if i.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(i.Columns) == 0 {
		return fmt.Errorf("columns is required")
	}
	return nil
}
