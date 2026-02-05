package tools

import (
	"encoding/json"

	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/domain/tools"
)

// EndJourneyTool completes the current workflow.
type EndJourneyTool struct{}

// NewEndJourneyTool creates a new end_journey tool.
func NewEndJourneyTool() *EndJourneyTool {
	return &EndJourneyTool{}
}

// Name returns the tool name.
func (t *EndJourneyTool) Name() string {
	return "end_journey"
}

// Definition returns the tool definition for the chat API.
func (t *EndJourneyTool) Definition() chat.Tool {
	return chat.Tool{
		Name:        t.Name(),
		Description: "Complete the current journey workflow",
		InputSchema: chat.NewObjectSchema(nil, nil),
	}
}

// Execute runs the tool and returns a typed result.
func (t *EndJourneyTool) Execute(input json.RawMessage) (tools.EndJourneyResult, error) {
	return tools.EndJourneyResult{Ended: true}, nil
}
