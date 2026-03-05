package tenancy

import (
	"context"
	"time"
)

type WorkspaceID string

type Workspace struct {
	ID        WorkspaceID
	AccountID AccountID
	Name      string
	CreatedAt time.Time
}

// WorkspaceService is the domain contract for workspace operations.
type WorkspaceService interface {
	Create(ctx context.Context, accountID AccountID, name string) (WorkspaceID, error)
	Delete(ctx context.Context, id WorkspaceID) error
	ListByAccount(ctx context.Context, accountID AccountID) ([]Workspace, error)
}
