package tools

import (
	"encoding/json"

	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/tui/app/tools/endjourney"
	"github.com/usetero/cli/internal/tui/app/tools/query"
	"github.com/usetero/cli/internal/tui/app/tools/startjourney"
)

// Tools holds the global tools available everywhere.
type Tools struct {
	StartJourney *startjourney.Tool
	EndJourney   *endjourney.Tool
	Query        *query.Tool
}

// Definitions returns chat.Tool definitions for the API request.
func (t Tools) Definitions() []chat.Tool {
	var defs []chat.Tool
	if t.StartJourney != nil {
		defs = append(defs, t.StartJourney.Definition())
	}
	if t.EndJourney != nil {
		defs = append(defs, t.EndJourney.Definition())
	}
	if t.Query != nil {
		defs = append(defs, t.Query.Definition())
	}
	return defs
}

// Execute runs a tool by name and returns the result.
// Returns nil, nil if tool not found.
func (t Tools) Execute(name string, input json.RawMessage) (any, error) {
	switch name {
	case startjourney.Name:
		if t.StartJourney != nil {
			return t.StartJourney.Execute(input)
		}
	case endjourney.Name:
		if t.EndJourney != nil {
			return t.EndJourney.Execute(input)
		}
	case query.Name:
		if t.Query != nil {
			return t.Query.Execute(input)
		}
	}
	return nil, nil
}

// Has returns true if a tool is available by name.
func (t Tools) Has(name string) bool {
	switch name {
	case startjourney.Name:
		return t.StartJourney != nil
	case endjourney.Name:
		return t.EndJourney != nil
	case query.Name:
		return t.Query != nil
	}
	return false
}
