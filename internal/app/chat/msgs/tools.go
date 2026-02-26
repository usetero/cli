package msgs

import (
	"github.com/usetero/cli/internal/domain"
	domaintools "github.com/usetero/cli/internal/domain/tools"
)

// ToolCompleted is fired when any tool finishes executing.
type ToolCompleted struct {
	TurnID    domain.MessageID
	ToolUseID string
	Result    domaintools.Result
	Error     error
}

// GetTurnID returns the turn ID.
func (m ToolCompleted) GetTurnID() domain.MessageID { return m.TurnID }

// GetToolUseID returns the tool use ID.
func (m ToolCompleted) GetToolUseID() string { return m.ToolUseID }

// GetError returns the execution error, if any.
func (m ToolCompleted) GetError() error { return m.Error }

// GetResult returns the tool result, wrapping the error if present.
func (m ToolCompleted) GetResult() domaintools.Result {
	r := m.Result
	r.ToolUseID = m.ToolUseID
	if m.Error != nil {
		r.Error = &domaintools.ErrorResult{Message: m.Error.Error()}
	}
	return r
}
