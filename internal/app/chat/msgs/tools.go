package msgs

import (
	domaintools "github.com/usetero/cli/internal/domain/tools"
)

// ToolCompleted is fired when any tool finishes executing.
type ToolCompleted struct {
	ToolUseID string
	Result    domaintools.Result
	Error     error
}

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

// StartPolicyApprovalCompleted is fired when the policy approval wizard is triggered.
type StartPolicyApprovalCompleted struct {
	ToolUseID string
	Started   bool
	Error     error
}

func (m StartPolicyApprovalCompleted) GetToolUseID() string { return m.ToolUseID }
func (m StartPolicyApprovalCompleted) GetError() error      { return m.Error }
func (m StartPolicyApprovalCompleted) GetResult() domaintools.Result {
	return domaintools.Result{
		ToolUseID:           m.ToolUseID,
		StartPolicyApproval: &domaintools.StartPolicyApprovalResult{Started: m.Started},
		Error:               errorResultFromErr(m.Error),
	}
}
func (m StartPolicyApprovalCompleted) toolCompleted() {}

func errorResultFromErr(err error) *domaintools.ErrorResult {
	if err == nil {
		return nil
	}
	return &domaintools.ErrorResult{Message: err.Error()}
}
