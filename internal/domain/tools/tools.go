// Package tools defines tool input/output types.
package tools

// Result holds a tool result with serialized content.
type Result struct {
	ToolUseID string
	Content   map[string]any // serialized tool result
	Error     *ErrorResult
}

// ToMap serializes the result for the GraphQL API.
func (r Result) ToMap() map[string]any {
	if r.Error != nil {
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
