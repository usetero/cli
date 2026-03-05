package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

type EnableServiceInput struct {
	ServiceID string `json:"service_id"`
	Enabled   bool   `json:"enabled"`
}

type EnableServiceResult struct {
	ServiceID ServiceID `json:"service_id"`
	Enabled   bool      `json:"enabled"`
}

type EnableServiceTool struct {
	enableFunc  func(ctx context.Context, serviceID ServiceID) error
	disableFunc func(ctx context.Context, serviceID ServiceID) error
}

func NewEnableServiceTool(
	enableFunc func(ctx context.Context, serviceID ServiceID) error,
	disableFunc func(ctx context.Context, serviceID ServiceID) error,
) *EnableServiceTool {
	return &EnableServiceTool{
		enableFunc:  enableFunc,
		disableFunc: disableFunc,
	}
}

func (t *EnableServiceTool) Definition() Definition {
	return Definition{
		Name:        EnableServiceToolName,
		Description: "Enable or disable service analysis.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"service_id": map[string]any{"type": "string"},
				"enabled":    map[string]any{"type": "boolean"},
			},
			"required": []string{"service_id", "enabled"},
		},
	}
}

func (t *EnableServiceTool) Run(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	if t == nil {
		return nil, fmt.Errorf("enable service tool is not initialized")
	}
	var in EnableServiceInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("parse enable service input: %w", err)
	}
	serviceID, err := ParseServiceID(in.ServiceID)
	if err != nil {
		return nil, err
	}

	if in.Enabled {
		if t.enableFunc == nil {
			return nil, fmt.Errorf("enable service function is not configured")
		}
		if err := t.enableFunc(ctx, serviceID); err != nil {
			return nil, err
		}
	} else {
		if t.disableFunc == nil {
			return nil, fmt.Errorf("disable service function is not configured")
		}
		if err := t.disableFunc(ctx, serviceID); err != nil {
			return nil, err
		}
	}

	return json.Marshal(EnableServiceResult{ServiceID: serviceID, Enabled: in.Enabled})
}
