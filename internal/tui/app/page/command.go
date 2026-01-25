package page

import "charm.land/bubbles/v2/key"

// Command represents a slash command that a page supports
type Command struct {
	Name        string // e.g. "sort", "filter"
	Description string // e.g. "Sort the table"
	Binding     key.Binding
}
