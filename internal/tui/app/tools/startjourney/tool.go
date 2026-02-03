package startjourney

import (
	"encoding/json"

	"github.com/usetero/cli/internal/chat"
)

// Tool begins a guided workflow.
type Tool struct{}

func (Tool) Definition() chat.Tool {
	return chat.Tool{
		Name:        "start_journey",
		Description: "Begin a guided workflow to help the user accomplish a goal",
		InputSchema: chat.Schema{
			Type: "object",
			Properties: map[string]chat.Property{
				"name": {
					Type:        "string",
					Description: "The journey to start",
					Enum:        []string{"first_encounter", "optimization"},
				},
			},
			Required: []string{"name"},
		},
	}
}

func (Tool) Execute(input json.RawMessage) (any, error) {
	return true, nil
}
