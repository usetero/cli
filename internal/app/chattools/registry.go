package tools

import (
	"encoding/json"

	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools/action"
	"github.com/usetero/cli/internal/boundary/chat"
	domaintools "github.com/usetero/cli/internal/domain/tools"
)

// ActionTool bundles a definition, executor, and display config for the generic action model.
type ActionTool struct {
	Def    chat.Tool
	Exec   action.Executor
	Config action.Config
}

// Registry holds tool instances and provides definitions.
type Registry struct {
	actions map[string]ActionTool
}

// NewRegistry creates a registry from a map of action tools. All chat tools
// are GraphQL-backed action tools; there is no local-SQL query surface.
func NewRegistry(actions map[string]ActionTool) *Registry {
	return &Registry{actions: actions}
}

// Definitions returns tool definitions for the chat API.
func (r *Registry) Definitions() []chat.Tool {
	defs := make([]chat.Tool, 0, len(r.actions))
	for _, a := range r.actions {
		defs = append(defs, a.Def)
	}
	return defs
}

// Lookup returns the action tool for the given name.
func (r *Registry) Lookup(name string) (ActionTool, bool) {
	a, ok := r.actions[name]
	return a, ok
}

// NewActionTool is a helper to build an ActionTool from parts.
func NewActionTool(
	def chat.Tool,
	executor func(input json.RawMessage) (domaintools.Result, error),
	config action.Config,
) ActionTool {
	return ActionTool{
		Def:    def,
		Exec:   executor,
		Config: config,
	}
}

// UnknownTool returns an ActionTool for tools not in the registry.
// It shows the tool name and completes with an error so the result
// is reported back to the model.
func UnknownTool(name string) ActionTool {
	return ActionTool{
		Exec: func(_ json.RawMessage) (domaintools.Result, error) {
			return domaintools.Result{
				Content: map[string]any{"error": "unknown tool"},
			}, nil
		},
		Config: action.Config{
			DisplayName: func(_ json.RawMessage) string { return name },
			Status:      func(_ json.RawMessage) string { return "Executing" },
			Result:      func(_ domaintools.Result) string { return "Unknown tool" },
		},
	}
}
