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
	if db == nil {
		panic("tenancy local workspace service requires db")
	}
	return &LocalWorkspaceService{q: workspacesdb.New(db)}
}

func (s *LocalWorkspaceService) Create(ctx context.Context, create WorkspaceCreate) (WorkspaceID, error) {
	validated, err := create.Validate()
	if err != nil {
		return "", err
	}

	id := WorkspaceID(uuid.NewString())
	err = s.q.Create(ctx, workspacesdb.CreateParams{
		ID:        toWorkspacesDBWorkspaceID(id),
		AccountID: toWorkspacesDBAccountID(validated.AccountID),
		Name:      validated.Name,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *LocalWorkspaceService) Delete(ctx context.Context, id WorkspaceID) error {
	if id == "" {
		return fmt.Errorf("workspace id is required")
	}
	return s.q.Delete(ctx, toWorkspacesDBWorkspaceID(id))
}

func (s *LocalWorkspaceService) ListByAccount(ctx context.Context, accountID AccountID) ([]Workspace, error) {
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
