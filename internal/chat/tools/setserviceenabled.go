package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/sqlite"
)

// SetServiceEnabledTool enables or disables a service.
type SetServiceEnabledTool struct {
	db sqlite.DB
}

// NewSetServiceEnabledTool creates a new set_service_enabled tool.
func NewSetServiceEnabledTool(db sqlite.DB) *SetServiceEnabledTool {
	return &SetServiceEnabledTool{db: db}
}

// Name returns the tool name.
func (t *SetServiceEnabledTool) Name() string {
	return "set_service_enabled"
}

// Definition returns the tool definition for the chat API.
func (t *SetServiceEnabledTool) Definition() chat.Tool {
	return chat.Tool{
		Name:        t.Name(),
		Description: "Enable or disable a service for log analysis. Enabling triggers the analysis pipeline.",
		InputSchema: chat.NewObjectSchema(
			map[string]chat.Property{
				"service_id": {
					Type:        "string",
					Description: "The ID of the service to enable or disable",
				},
				"enabled": {
					Type:        "boolean",
					Description: "true to enable, false to disable",
				},
			},
			[]string{"service_id", "enabled"},
		),
	}
}

// Execute runs the tool and returns a typed result.
func (t *SetServiceEnabledTool) Execute(input json.RawMessage) (tools.SetServiceEnabledResult, error) {
	var in tools.SetServiceEnabledInput
	if err := json.Unmarshal(input, &in); err != nil {
		return tools.SetServiceEnabledResult{}, err
	}

	ctx := context.Background()

	// Write to local SQLite — PowerSync will sync this to the server via the upload handler.
	if err := t.db.Services().SetEnabled(ctx, in.ServiceID, in.Enabled); err != nil {
		return tools.SetServiceEnabledResult{}, fmt.Errorf("set service enabled: %w", err)
	}

	// Read back to get service name for the result.
	var serviceName string
	row := t.db.QueryRow(ctx, "SELECT name FROM services WHERE id = ?", in.ServiceID)
	if err := row.Scan(&serviceName); err != nil {
		serviceName = in.ServiceID // fall back to ID if name unavailable
	}

	return tools.SetServiceEnabledResult{
		ServiceID:   in.ServiceID,
		ServiceName: serviceName,
		Enabled:     in.Enabled,
	}, nil
}
