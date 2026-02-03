package endjourney

import (
	"encoding/json"

	"github.com/usetero/cli/internal/chat"
)

// Tool completes the current workflow.
type Tool struct{}

func (Tool) Definition() chat.Tool {
	return chat.Tool{
		Name:        "end_journey",
		Description: "Complete the current journey workflow",
		InputSchema: chat.Schema{
			Type: "object",
		},
	}
}

func (Tool) Execute(input json.RawMessage) (any, error) {
	return true, nil
}
