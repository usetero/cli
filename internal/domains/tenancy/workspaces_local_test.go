package tenancy

import (
	"context"
	"strings"
	"testing"
)

func TestLocalWorkspaceService_CRUD(t *testing.T) {
	db := openTenancyTestDB(t)
	accountSvc := NewLocalAccountService(db.Raw())
	workspaceSvc := NewLocalWorkspaceService(db.Raw())

	accountID, err := accountSvc.Create(context.Background(), AccountCreate{Name: "Primary"})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	workspaceID, err := workspaceSvc.Create(context.Background(), WorkspaceCreate{AccountID: accountID, Name: "Default"})
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if workspaceID == "" {
		t.Fatalf("expected non-empty workspace id")
	}

	workspaces, err := workspaceSvc.ListByAccount(context.Background(), accountID)
	if err != nil {
		t.Fatalf("list workspaces: %v", err)
	}
	if len(workspaces) != 1 {
		t.Fatalf("expected one workspace, got %d", len(workspaces))
	}
	if workspaces[0].ID != workspaceID || workspaces[0].AccountID != accountID || workspaces[0].Name != "Default" {
		t.Fatalf("unexpected workspace row: %+v", workspaces[0])
	}

	if err := workspaceSvc.Delete(context.Background(), workspaceID); err != nil {
		t.Fatalf("delete workspace: %v", err)
	}
	workspaces, err = workspaceSvc.ListByAccount(context.Background(), accountID)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(workspaces) != 0 {
		t.Fatalf("expected no workspaces after delete, got %d", len(workspaces))
	}
}

func TestLocalWorkspaceService_ValidationAndUninitialized(t *testing.T) {
	db := openTenancyTestDB(t)
	svc := NewLocalWorkspaceService(db.Raw())

	if _, err := svc.Create(context.Background(), WorkspaceCreate{Name: "Default"}); err == nil || !strings.Contains(err.Error(), "account id is required") {
		t.Fatalf("expected account id validation error, got %v", err)
	}
	if _, err := svc.Create(context.Background(), WorkspaceCreate{AccountID: "acc_1"}); err == nil || !strings.Contains(err.Error(), "workspace name is required") {
		t.Fatalf("expected workspace name validation error, got %v", err)
	}
	if _, err := svc.ListByAccount(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "account id is required") {
		t.Fatalf("expected account id validation on list, got %v", err)
	}
	if err := svc.Delete(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "workspace id is required") {
		t.Fatalf("expected workspace id validation error, got %v", err)
	}
}
