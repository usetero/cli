package msgs

import (
	domaintools "github.com/usetero/cli/internal/domain/tools"
)

// ToolCompleted is the interface for all tool completion messages.
// The unexported marker method ensures only intentional implementations satisfy it.
type ToolCompleted interface {
	GetToolUseID() string
	GetError() error
	GetResult() domaintools.Result
	toolCompleted() // marker method
}

// QueryCompleted is fired when a query tool finishes executing.
type QueryCompleted struct {
	ToolUseID string
	Result    domaintools.QueryResult
	Error     error
}

func (m QueryCompleted) GetToolUseID() string { return m.ToolUseID }
func (m QueryCompleted) GetError() error      { return m.Error }
func (m QueryCompleted) GetResult() domaintools.Result {
	return domaintools.Result{
		ToolUseID: m.ToolUseID,
		Query:     &m.Result,
		Error:     errorResultFromErr(m.Error),
	}
}
func (m QueryCompleted) toolCompleted() {}

// StartJourneyCompleted is fired when a start_journey tool finishes executing.
type StartJourneyCompleted struct {
	ToolUseID string
	Result    domaintools.StartJourneyResult
	Error     error
}

func (m StartJourneyCompleted) GetToolUseID() string { return m.ToolUseID }
func (m StartJourneyCompleted) GetError() error      { return m.Error }
func (m StartJourneyCompleted) GetResult() domaintools.Result {
	return domaintools.Result{
		ToolUseID:    m.ToolUseID,
		StartJourney: &m.Result,
		Error:        errorResultFromErr(m.Error),
	}
}
func (m StartJourneyCompleted) toolCompleted() {}

// EndJourneyCompleted is fired when an end_journey tool finishes executing.
type EndJourneyCompleted struct {
	ToolUseID string
	Result    domaintools.EndJourneyResult
	Error     error
}

func (m EndJourneyCompleted) GetToolUseID() string { return m.ToolUseID }
func (m EndJourneyCompleted) GetError() error      { return m.Error }
func (m EndJourneyCompleted) GetResult() domaintools.Result {
	return domaintools.Result{
		ToolUseID:  m.ToolUseID,
		EndJourney: &m.Result,
		Error:      errorResultFromErr(m.Error),
	}
}
func (m EndJourneyCompleted) toolCompleted() {}

// SetServiceEnabledCompleted is fired when a set_service_enabled tool finishes executing.
type SetServiceEnabledCompleted struct {
	ToolUseID string
	Result    domaintools.SetServiceEnabledResult
	Error     error
}

func (m SetServiceEnabledCompleted) GetToolUseID() string { return m.ToolUseID }
func (m SetServiceEnabledCompleted) GetError() error      { return m.Error }
func (m SetServiceEnabledCompleted) GetResult() domaintools.Result {
	return domaintools.Result{
		ToolUseID:         m.ToolUseID,
		SetServiceEnabled: &m.Result,
		Error:             errorResultFromErr(m.Error),
	}
}
func (m SetServiceEnabledCompleted) toolCompleted() {}

func errorResultFromErr(err error) *domaintools.ErrorResult {
	if err == nil {
		return nil
	}
	return &domaintools.ErrorResult{Message: err.Error()}
}

// Compile-time interface checks
var (
	_ ToolCompleted = QueryCompleted{}
	_ ToolCompleted = StartJourneyCompleted{}
	_ ToolCompleted = EndJourneyCompleted{}
	_ ToolCompleted = SetServiceEnabledCompleted{}
)
