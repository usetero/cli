package tools

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools/action"
	"github.com/usetero/cli/internal/boundary/chat"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/domain/tools"
)

// NewSetServiceEnabledAction creates an ActionTool for set_service_enabled.
// The enable/disable write is a synchronous control-plane GraphQL mutation.
func NewSetServiceEnabledAction(services graphql.Services) ActionTool {
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

		if _, parseErr := uuid.Parse(in.ServiceID.String()); parseErr != nil {
			return tools.Result{}, fmt.Errorf(
				"service ID %q is not a UUID — this looks like a name. "+
					"Use the query tool: SELECT id, name FROM services WHERE name LIKE '%%%s%%'",
				in.ServiceID, in.ServiceID,
			)
		}

		ctx, cancel := withToolTimeout()
		defer cancel()

		var err error
		if in.Enabled {
			err = services.EnableService(ctx, in.ServiceID)
		} else {
			err = services.DisableService(ctx, in.ServiceID)
		}
		if err != nil {
			return tools.Result{}, fmt.Errorf("set service enabled: %w", err)
		}

		return tools.Result{
			Content: tools.SetServiceEnabledResult{
				ServiceID: in.ServiceID,
				Enabled:   in.Enabled,
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
