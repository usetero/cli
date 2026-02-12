package tools

import "github.com/usetero/cli/internal/chat"

// Registry holds tool instances and provides definitions.
type Registry struct {
	Query             *QueryTool
	StartJourney      *StartJourneyTool
	EndJourney        *EndJourneyTool
	SetServiceEnabled *SetServiceEnabledTool
}

// Definitions returns tool definitions for the chat API.
func (r *Registry) Definitions() []chat.Tool {
	var defs []chat.Tool
	if r.Query != nil {
		defs = append(defs, r.Query.Definition())
	}
	if r.StartJourney != nil {
		defs = append(defs, r.StartJourney.Definition())
	}
	if r.EndJourney != nil {
		defs = append(defs, r.EndJourney.Definition())
	}
	if r.SetServiceEnabled != nil {
		defs = append(defs, r.SetServiceEnabled.Definition())
	}
	return defs
}
