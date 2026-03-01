package events

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

// ResultOrError returns the tool result, wrapping Error when present.
func (m ToolCompleted) ResultOrError() domaintools.Result {
	r := m.Result
	r.ToolUseID = m.ToolUseID
	if m.Error != nil {
		r.Error = &domaintools.ErrorResult{Message: m.Error.Error()}
	}
	return r
}
