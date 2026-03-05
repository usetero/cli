package tenancy

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	workspacesdb "github.com/usetero/cli/internal/domains/tenancy/db/workspacesgen"
)

// LocalWorkspaceService uses SQLite/sqlc for workspace CRUD.
type LocalWorkspaceService struct {
	q *workspacesdb.Queries
}

func NewLocalWorkspaceService(db *sql.DB) *LocalWorkspaceService {
	return &LocalWorkspaceService{q: workspacesdb.New(db)}
}

func (s *LocalWorkspaceService) Create(ctx context.Context, accountID AccountID, name string) (WorkspaceID, error) {
	if s == nil || s.q == nil {
		return "", fmt.Errorf("tenancy local workspace service is not initialized")
	}
	if accountID == "" {
		return "", fmt.Errorf("account id is required")
	}
	if name == "" {
		return "", fmt.Errorf("workspace name is required")
	}

	id := WorkspaceID(uuid.NewString())
	err := s.q.Create(ctx, workspacesdb.CreateParams{
		ID:        toWorkspacesDBWorkspaceID(id),
		AccountID: toWorkspacesDBAccountID(accountID),
		Name:      name,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *LocalWorkspaceService) Delete(ctx context.Context, id WorkspaceID) error {
	if s == nil || s.q == nil {
		return fmt.Errorf("tenancy local workspace service is not initialized")
	}
	if id == "" {
		return fmt.Errorf("workspace id is required")
	}
	return s.q.Delete(ctx, toWorkspacesDBWorkspaceID(id))
}

func (s *LocalWorkspaceService) ListByAccount(ctx context.Context, accountID AccountID) ([]Workspace, error) {
	if s == nil || s.q == nil {
		return nil, fmt.Errorf("tenancy local workspace service is not initialized")
	}
	if accountID == "" {
		return nil, fmt.Errorf("account id is required")
	}
	rows, err := s.q.ListByAccount(ctx, toWorkspacesDBAccountID(accountID))
	if err != nil {
		return nil, err
	}

	out := make([]Workspace, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromWorkspacesDBWorkspace(row))
	}
	return out, nil
}
