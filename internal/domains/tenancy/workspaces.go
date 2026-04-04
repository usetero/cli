package tenancy

import (
	"context"
	"strings"
	"time"

	"github.com/usetero/cli/internal/domains/validation"
)

type WorkspaceID string

type Workspace struct {
	ID        WorkspaceID
	AccountID AccountID
	Name      string
	CreatedAt time.Time
}

// WorkspaceCreate is the workspace creation mutation input.
type WorkspaceCreate struct {
	Name string `label:"workspace name" validate:"required,notblank,max=100"`
}

// Validate normalizes and validates workspace create input.
func (c WorkspaceCreate) Validate() (WorkspaceCreate, error) {
	c.Name = strings.TrimSpace(c.Name)
	if err := validation.Struct(c); err != nil {
		return WorkspaceCreate{}, err
	}
	return c, nil
}

// WorkspaceService is the domain contract for workspace operations.
type WorkspaceService interface {
	Create(ctx context.Context, create WorkspaceCreate) (WorkspaceID, error)
	Delete(ctx context.Context, id WorkspaceID) error
	List(ctx context.Context) ([]Workspace, error)
}
