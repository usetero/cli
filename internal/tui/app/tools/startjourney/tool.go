package startjourney

import (
	"encoding/json"

	"github.com/usetero/cli/internal/chat"
)

// Name is the tool name used in definitions and lookups.
const Name = "start_journey"

// Tool begins a guided workflow.
type Tool struct{}

func (Tool) Definition() chat.Tool {
	return chat.Tool{
		Name:        Name,
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

func (Tool) Execute(input json.RawMessage) (any, error) {
	return true, nil
}
