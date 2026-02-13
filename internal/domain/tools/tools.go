// Package tools defines tool input/output types.
package tools

// Result holds a tool result with serialized content.
type Result struct {
	ToolUseID           string
	Content             map[string]any // serialized tool result (for generic action tools)
	Query               *QueryResult
	StartPolicyApproval *StartPolicyApprovalResult
	StartJourney        *StartJourneyResult
	EndJourney          *EndJourneyResult
	Error               *ErrorResult
}

// ToMap serializes the result for the GraphQL API.
func (r Result) ToMap() map[string]any {
	switch {
	case r.Query != nil:
		return r.Query.ToMap()
	case r.StartPolicyApproval != nil:
		return r.StartPolicyApproval.ToMap()
	case r.StartJourney != nil:
		return r.StartJourney.ToMap()
	case r.EndJourney != nil:
		return r.EndJourney.ToMap()
	case r.Error != nil:
		return map[string]any{"error": r.Error.Message}
	}
	return r.Content
}

// IsError returns true if this result represents an error.
func (r Result) IsError() bool {
	return r.Error != nil
}

// ErrorResult represents a tool execution error.
type ErrorResult struct {
	Message string
}
