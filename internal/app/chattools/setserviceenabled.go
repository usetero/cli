package tools

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools/action"
	"github.com/usetero/cli/internal/boundary/chat"
	"github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/sqlite"
)

// NewSetServiceEnabledAction creates an ActionTool for set_service_enabled.
func NewSetServiceEnabledAction(db sqlite.DB) ActionTool {
	def := chat.Tool{
		Name:        "set_service_enabled",
		Description: "Enable or disable a service for log analysis. Enabling triggers the analysis pipeline.",
		InputSchema: chat.NewObjectSchema(
			map[string]chat.Property{
				"service_id": {
					Type:        "string",
					Description: "UUID of the service (e.g., '4a3b1c2d-...'). Use the query tool to look up service IDs: SELECT id, name FROM services",
				},
				"enabled": {
					Type:        "boolean",
					Description: "true to enable, false to disable",
				},
			},
			[]string{"service_id", "enabled"},
		),
	}

	executor := func(input json.RawMessage) (tools.Result, error) {
		var in tools.SetServiceEnabledInput
		if err := json.Unmarshal(input, &in); err != nil {
			return tools.Result{}, err
		}

		ctx, cancel := withToolTimeout()
		defer cancel()

		svc, err := db.Services().Get(ctx, in.ServiceID)
		if errors.Is(err, sql.ErrNoRows) {
			if _, parseErr := uuid.Parse(in.ServiceID.String()); parseErr != nil {
				return tools.Result{}, fmt.Errorf(
					"no service found with ID %q — this looks like a name, not a UUID. "+
						"Use the query tool: SELECT id, name FROM services WHERE name LIKE '%%%s%%'",
					in.ServiceID, in.ServiceID,
				)
			}
			return tools.Result{}, fmt.Errorf(
				"no service found with ID %q. Use the query tool: SELECT id, name FROM services",
				in.ServiceID,
			)
		}
		if err != nil {
			return tools.Result{}, fmt.Errorf("get service: %w", err)
		}

		if err := db.Services().SetEnabled(ctx, in.ServiceID, in.Enabled); err != nil {
			return tools.Result{}, fmt.Errorf("set service enabled: %w", err)
		}

		var serviceName string
		if svc.Name != nil {
			serviceName = *svc.Name
		}

		return tools.Result{
			Content: tools.SetServiceEnabledResult{
				ServiceID:   in.ServiceID,
				ServiceName: serviceName,
				Enabled:     in.Enabled,
			}.ToMap(),
		}, nil
	}

	config := action.Config{
		DisplayName: func(input json.RawMessage) string {
			var in tools.SetServiceEnabledInput
			if json.Unmarshal(input, &in) == nil && !in.Enabled {
				return "Disable Service"
			}
			return "Enable Service"
		},
		Status: func(input json.RawMessage) string {
			var in tools.SetServiceEnabledInput
			if json.Unmarshal(input, &in) != nil {
				return ""
			}
			if in.Enabled {
				return fmt.Sprintf("Enabling %s", in.ServiceID)
			}
			return fmt.Sprintf("Disabling %s", in.ServiceID)
		},
		Result: func(result tools.Result) string {
			name, _ := result.Content["service_name"].(string)
			if name == "" {
				name, _ = result.Content["service_id"].(string)
			}
			enabled, _ := result.Content["enabled"].(bool)
			if enabled {
				return fmt.Sprintf("%s enabled", name)
			}
			return fmt.Sprintf("%s disabled", name)
		},
	}

	return NewActionTool(def, executor, config)
}
