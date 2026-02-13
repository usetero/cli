package upload

import (
	"context"
	"errors"
	"testing"

	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync/db"
)

func TestServiceHandler_Handle(t *testing.T) {
	t.Parallel()

	t.Run("PATCH enabled=true calls EnableService", func(t *testing.T) {
		t.Parallel()

		var enabledID domain.ServiceID
		mock := &apitest.MockAPIServiceServices{
			EnableServiceFunc: func(ctx context.Context, serviceID domain.ServiceID) error {
				enabledID = serviceID
				return nil
			},
		}

		h := newServiceHandler(mock, logtest.NewScope(t))
		entry := &db.CrudEntry{
			Op:    db.OpPatch,
			RowID: "svc-1",
			Data:  map[string]any{"enabled": true},
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}
		if enabledID != "svc-1" {
			t.Errorf("EnableService called with %q, want %q", enabledID, "svc-1")
		}
	})

	t.Run("PATCH enabled=false calls DisableService", func(t *testing.T) {
		t.Parallel()

		var disabledID domain.ServiceID
		mock := &apitest.MockAPIServiceServices{
			DisableServiceFunc: func(ctx context.Context, serviceID domain.ServiceID) error {
				disabledID = serviceID
				return nil
			},
		}

		h := newServiceHandler(mock, logtest.NewScope(t))
		entry := &db.CrudEntry{
			Op:    db.OpPatch,
			RowID: "svc-2",
			Data:  map[string]any{"enabled": false},
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}
		if disabledID != "svc-2" {
			t.Errorf("DisableService called with %q, want %q", disabledID, "svc-2")
		}
	})

	t.Run("PATCH enabled as int 1 calls EnableService", func(t *testing.T) {
		t.Parallel()

		var called bool
		mock := &apitest.MockAPIServiceServices{
			EnableServiceFunc: func(ctx context.Context, serviceID domain.ServiceID) error {
				called = true
				return nil
			},
		}

		h := newServiceHandler(mock, logtest.NewScope(t))
		entry := &db.CrudEntry{
			Op:    db.OpPatch,
			RowID: "svc-3",
			Data:  map[string]any{"enabled": float64(1)}, // JSON decodes numbers as float64
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}
		if !called {
			t.Error("expected EnableService to be called")
		}
	})

	t.Run("PATCH returns error on API failure", func(t *testing.T) {
		t.Parallel()

		mock := &apitest.MockAPIServiceServices{
			EnableServiceFunc: func(ctx context.Context, serviceID domain.ServiceID) error {
				return errors.New("network error")
			},
		}

		h := newServiceHandler(mock, logtest.NewScope(t))
		entry := &db.CrudEntry{
			Op:    db.OpPatch,
			RowID: "svc-1",
			Data:  map[string]any{"enabled": true},
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err == nil {
			t.Error("Handle() expected error, got nil")
		}
	})

	t.Run("PATCH without enabled field drops silently", func(t *testing.T) {
		t.Parallel()

		mock := apitest.NewMockAPIServiceServices()
		h := newServiceHandler(mock, logtest.NewScope(t))
		entry := &db.CrudEntry{
			Op:    db.OpPatch,
			RowID: "svc-1",
			Data:  map[string]any{"name": "updated-name"},
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Errorf("Handle() error = %v, want nil for unhandled fields", err)
		}
	})

	t.Run("PUT drops silently", func(t *testing.T) {
		t.Parallel()

		mock := apitest.NewMockAPIServiceServices()
		h := newServiceHandler(mock, logtest.NewScope(t))
		entry := &db.CrudEntry{
			Op:    db.OpPut,
			RowID: "svc-1",
			Data:  map[string]any{},
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Errorf("Handle() error = %v, want nil for unsupported op", err)
		}
	})

	t.Run("DELETE drops silently", func(t *testing.T) {
		t.Parallel()

		mock := apitest.NewMockAPIServiceServices()
		h := newServiceHandler(mock, logtest.NewScope(t))
		entry := &db.CrudEntry{
			Op:    db.OpDelete,
			RowID: "svc-1",
			Data:  map[string]any{},
		}

		err := h.Handle(context.Background(), entry, noopEmitter())
		if err != nil {
			t.Errorf("Handle() error = %v, want nil for unsupported op", err)
		}
	})
}
