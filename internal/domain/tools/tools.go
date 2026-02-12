// Package tools defines tool input/output types and the execution interface.
package tools

import "encoding/json"

// Tool is the interface for executable tools.
type Tool interface {
	Name() string
	Execute(input json.RawMessage) (Result, error)
}

// Result holds a typed tool result. Exactly one field is set.
type Result struct {
	ToolUseID         string
	Query             *QueryResult
	StartJourney      *StartJourneyResult
	EndJourney        *EndJourneyResult
	SetServiceEnabled *SetServiceEnabledResult
	Error             *ErrorResult
}

// ToMap serializes the result for the GraphQL API.
func (r Result) ToMap() map[string]any {
	switch {
	case r.Query != nil:
		return r.Query.ToMap()
	case r.StartJourney != nil:
		return r.StartJourney.ToMap()
	case r.EndJourney != nil:
		return r.EndJourney.ToMap()
	case r.SetServiceEnabled != nil:
		return r.SetServiceEnabled.ToMap()
	case r.Error != nil:
		return map[string]any{"error": r.Error.Message}
	default:
		return nil
	}
}

// IsError returns true if this result represents an error.
func (r Result) IsError() bool {
	return r.Error != nil
}

// ErrorResult represents a tool execution error.
type ErrorResult struct {
	Message string
}
