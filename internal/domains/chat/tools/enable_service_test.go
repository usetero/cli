package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestEnableServiceTool_Run(t *testing.T) {
	t.Parallel()

	enabled := ServiceID("")
	disabled := ServiceID("")
	tool := NewEnableServiceTool(
		func(_ context.Context, serviceID ServiceID) error {
			enabled = serviceID
			return nil
		},
		func(_ context.Context, serviceID ServiceID) error {
			disabled = serviceID
			return nil
		},
	)

	out, err := tool.Run(context.Background(), json.RawMessage(`{"service_id":"svc_1","enabled":true}`))
	if err != nil {
		t.Fatalf("run enable: %v", err)
	}
	if enabled != "svc_1" || disabled != "" {
		t.Fatalf("unexpected function calls enabled=%q disabled=%q", enabled, disabled)
	}
	var parsed EnableServiceResult
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("parse output: %v", err)
	}
	if !parsed.Enabled {
		t.Fatalf("expected enabled output")
	}
}

func TestEnableServiceTool_RejectsMissingServiceID(t *testing.T) {
	t.Parallel()

	tool := NewEnableServiceTool(func(context.Context, ServiceID) error { return nil }, func(context.Context, ServiceID) error { return nil })
	_, err := tool.Run(context.Background(), json.RawMessage(`{"enabled":true}`))
	if err == nil || !strings.Contains(err.Error(), "service_id is required") {
		t.Fatalf("expected missing service_id error, got %v", err)
	}
}
