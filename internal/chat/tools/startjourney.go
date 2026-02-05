package tools

import (
	"encoding/json"

	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/domain/tools"
)

// StartJourneyTool begins a guided workflow.
type StartJourneyTool struct{}

// NewStartJourneyTool creates a new start_journey tool.
func NewStartJourneyTool() *StartJourneyTool {
	return &StartJourneyTool{}
}

// Name returns the tool name.
func (t *StartJourneyTool) Name() string {
	return "start_journey"
}

// Definition returns the tool definition for the chat API.
func (t *StartJourneyTool) Definition() chat.Tool {
	return chat.Tool{
		Name:        t.Name(),
		Description: "Begin a guided workflow to help the user accomplish a goal",
		InputSchema: chat.NewObjectSchema(
			map[string]chat.Property{
				"name": {
					Type:        "string",
					Description: "The journey to start",
					Enum:        []string{"first_encounter", "optimization"},
				},
			},
			[]string{"name"},
		),
	}
}

// Execute runs the tool and returns a typed result.
func (t *StartJourneyTool) Execute(input json.RawMessage) (tools.StartJourneyResult, error) {
	return tools.StartJourneyResult{Started: true}, nil
}
