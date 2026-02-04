package endjourney

import (
	"encoding/json"

	"github.com/usetero/cli/internal/chat"
)

// Name is the tool name used in definitions and lookups.
const Name = "end_journey"

// Tool completes the current workflow.
type Tool struct{}

func (Tool) Definition() chat.Tool {
	return chat.Tool{
		Name:        Name,
		Description: "Complete the current journey workflow",
		InputSchema: chat.NewObjectSchema(nil, nil),
	}
}

func (Tool) Execute(input json.RawMessage) (any, error) {
	return true, nil
}
