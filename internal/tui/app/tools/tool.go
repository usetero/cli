package tools

import (
	"encoding/json"

	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/tui/app/tools/endjourney"
	"github.com/usetero/cli/internal/tui/app/tools/startjourney"
)

// Tool defines a tool the AI can call.
type Tool interface {
	Definition() chat.Tool
	Execute(input json.RawMessage) (any, error)
}

// Tools is a collection of tools.
type Tools []Tool

// Definitions returns chat.Tool definitions for the request.
func (t Tools) Definitions() []chat.Tool {
	defs := make([]chat.Tool, len(t))
	for i, tool := range t {
		defs[i] = tool.Definition()
	}
	return defs
}

// Merge combines two tool sets.
func (t Tools) Merge(other Tools) Tools {
	return append(t, other...)
}

// Get returns a tool by name, or nil if not found.
func (t Tools) Get(name string) Tool {
	for _, tool := range t {
		if tool.Definition().Name == name {
			return tool
		}
	}
	return nil
}

// All returns all global tools.
func All() Tools {
	return Tools{
		startjourney.Tool{},
		endjourney.Tool{},
	}
}
