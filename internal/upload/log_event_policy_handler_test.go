package upload

import (
	"context"
	"errors"
	"testing"

	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync/db"
)

// mockPolicies implements api.LogEventPolicies for testing.
type mockPolicies struct {
	approveFunc func(ctx context.Context, id string) error
	dismissFunc func(ctx context.Context, id string) error
}

func (m *mockPolicies) Approve(ctx context.Context, id string) error {
	if m.approveFunc != nil {
		return m.approveFunc(ctx, id)
	}
	return nil
}

func (m *mockPolicies) Dismiss(ctx context.Context, id string) error {
	if m.dismissFunc != nil {
		return m.dismissFunc(ctx, id)
	}
	return nil
}

func TestPolicyHandler_Handle(t *testing.T) {
	t.Parallel()

	t.Run("PATCH with approved_at calls Approve", func(t *testing.T) {
		t.Parallel()

		var approvedID string
		policies := &mockPolicies{
			approveFunc: func(ctx context.Context, id string) error {
				approvedID = id
				return nil
			},
		}

		handler := newPolicyHandler(policies, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Table: "log_event_policies",
			RowID: "policy-123",
			Op:    db.OpPatch,
			Data: map[string]any{
				"approved_at": "2024-01-15T10:00:00Z",
				"approved_by": "user-456",
			},
		}

		err := handler.Handle(context.Background(), entry, nil)
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if approvedID != "policy-123" {
			t.Errorf("Approve called with id = %q, want %q", approvedID, "policy-123")
		}
	})

	t.Run("PATCH with dismissed_at calls Dismiss", func(t *testing.T) {
		t.Parallel()

		var dismissedID string
		policies := &mockPolicies{
			dismissFunc: func(ctx context.Context, id string) error {
				dismissedID = id
				return nil
			},
		}

		handler := newPolicyHandler(policies, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Table: "log_event_policies",
			RowID: "policy-789",
			Op:    db.OpPatch,
			Data: map[string]any{
				"dismissed_at": "2024-01-15T10:00:00Z",
				"dismissed_by": "user-456",
			},
		}

		err := handler.Handle(context.Background(), entry, nil)
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if dismissedID != "policy-789" {
			t.Errorf("Dismiss called with id = %q, want %q", dismissedID, "policy-789")
		}
	})

	t.Run("PATCH returns error on Approve failure", func(t *testing.T) {
		t.Parallel()

		policies := &mockPolicies{
			approveFunc: func(ctx context.Context, id string) error {
				return errors.New("api error")
			},
		}

		handler := newPolicyHandler(policies, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Table: "log_event_policies",
			RowID: "policy-123",
			Op:    db.OpPatch,
			Data: map[string]any{
				"approved_at": "2024-01-15T10:00:00Z",
			},
		}

		err := handler.Handle(context.Background(), entry, nil)
		if err == nil {
			t.Fatal("Handle() expected error, got nil")
		}
	})

	t.Run("ignores PUT operations", func(t *testing.T) {
		t.Parallel()

		policies := &mockPolicies{
			approveFunc: func(ctx context.Context, id string) error {
				t.Error("Approve should not be called")
				return nil
			},
		}

		handler := newPolicyHandler(policies, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Table: "log_event_policies",
			RowID: "policy-123",
			Op:    db.OpPut,
			Data:  map[string]any{},
		}

		err := handler.Handle(context.Background(), entry, nil)
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}
	})

	t.Run("ignores PATCH without approve/dismiss data", func(t *testing.T) {
		t.Parallel()

		policies := &mockPolicies{
			approveFunc: func(ctx context.Context, id string) error {
				t.Error("Approve should not be called")
				return nil
			},
			dismissFunc: func(ctx context.Context, id string) error {
				t.Error("Dismiss should not be called")
				return nil
			},
		}

		handler := newPolicyHandler(policies, logtest.NewScope(t))

		entry := &db.CrudEntry{
			Table: "log_event_policies",
			RowID: "policy-123",
			Op:    db.OpPatch,
			Data: map[string]any{
				"some_other_field": "value",
			},
		}

		err := handler.Handle(context.Background(), entry, nil)
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}
	})
}
