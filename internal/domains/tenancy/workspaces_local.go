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
	q         *workspacesdb.Queries
	accountID AccountID
}

func NewLocalWorkspaceService(db *sql.DB, accountID AccountID) *LocalWorkspaceService {
	if db == nil {
		panic("tenancy local workspace service requires db")
	}
	if accountID == "" {
		panic("tenancy local workspace service requires account id")
	}
	return &LocalWorkspaceService{q: workspacesdb.New(db), accountID: accountID}
}

func (s *LocalWorkspaceService) Create(ctx context.Context, create WorkspaceCreate) (WorkspaceID, error) {
	validated, err := create.Validate()
	if err != nil {
		return "", err
	}

	id := WorkspaceID(uuid.NewString())
	err = s.q.Create(ctx, workspacesdb.CreateParams{
		ID:        ptrString(toWorkspacesDBWorkspaceID(id)),
		AccountID: ptrString(toWorkspacesDBAccountID(s.accountID)),
		Name:      ptrString(validated.Name),
		CreatedAt: ptrString(time.Now().UTC().Format(time.RFC3339)),
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
	return s.q.Delete(ctx, ptrString(toWorkspacesDBWorkspaceID(id)))
}

func (s *LocalWorkspaceService) List(ctx context.Context) ([]Workspace, error) {
	rows, err := s.q.ListByAccount(ctx, ptrString(toWorkspacesDBAccountID(s.accountID)))
	if err != nil {
		return nil, err
	}

	out := make([]Workspace, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromWorkspacesDBListRow(row))
	}
	return out, nil
}
