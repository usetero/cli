package tenancy

import (
	"context"
	"strings"
	"testing"
	"time"

	controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"
	"github.com/usetero/cli/internal/infrastructure/controlplane/api/apitest"
)

func TestRemoteWorkspaceService_MappingAndValidation(t *testing.T) {
	svc := NewRemoteWorkspaceService(&apitest.Client{})
	if _, err := svc.ListByAccount(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "account id is required") {
		t.Fatalf("expected account id validation error, got %v", err)
	}

	var calledAccountID controlplane.AccountID
	var calledDeleteID controlplane.WorkspaceID
	now := time.Now().UTC()
	mock := &apitest.Client{
		ListWorkspacesFn: func(_ context.Context, accountID controlplane.AccountID) ([]controlplane.Workspace, error) {
			calledAccountID = accountID
			return []controlplane.Workspace{{ID: "ws_1", Name: "Default", CreatedAt: now}}, nil
		},
		DeleteWorkspaceFn: func(_ context.Context, workspaceID controlplane.WorkspaceID) error {
			calledDeleteID = workspaceID
			return nil
		},
	}
	svc = NewRemoteWorkspaceService(mock)
	rows, err := svc.ListByAccount(context.Background(), "acc_1")
	if err != nil {
		t.Fatalf("list workspaces: %v", err)
	}
	if calledAccountID != controlplane.AccountID("acc_1") {
		t.Fatalf("list called with account=%q", calledAccountID)
	}
	if len(rows) != 1 || rows[0].ID != "ws_1" || rows[0].AccountID != "acc_1" {
		t.Fatalf("unexpected workspace mapping: %+v", rows)
	}

	if _, err := svc.Create(context.Background(), WorkspaceCreate{AccountID: "acc_1", Name: "Default"}); err == nil || !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("expected not implemented create error, got %v", err)
	}
	if err := svc.Delete(context.Background(), "ws_1"); err != nil {
		t.Fatalf("delete workspace: %v", err)
	}
	if calledDeleteID != controlplane.WorkspaceID("ws_1") {
		t.Fatalf("delete called with workspace=%q", calledDeleteID)
	}
	if err := svc.Delete(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "workspace id is required") {
		t.Fatalf("expected workspace id validation error, got %v", err)
	}
}
