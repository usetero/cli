package tenancy

import (
	"context"
	"strings"
	"time"

	"github.com/usetero/cli/internal/domains/validation"
)

type AccountID string

type Account struct {
	ID        AccountID
	Name      string
	CreatedAt time.Time
}

// AccountCreate is the account creation mutation input.
type AccountCreate struct {
	Name string `label:"account name" validate:"required,notblank,max=100"`
}

// Validate normalizes and validates account create input.
func (c AccountCreate) Validate() (AccountCreate, error) {
	c.Name = strings.TrimSpace(c.Name)
	if err := validation.Struct(c); err != nil {
		return AccountCreate{}, err
	}
	return c, nil
}

// AccountService is the domain contract for account operations.
type AccountService interface {
	Create(ctx context.Context, create AccountCreate) (AccountID, error)
	Delete(ctx context.Context, id AccountID) error
	List(ctx context.Context) ([]Account, error)
}
