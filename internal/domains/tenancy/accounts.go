package tenancy

import (
	"context"
	"time"
)

type AccountID string

type Account struct {
	ID        AccountID
	Name      string
	CreatedAt time.Time
}

// AccountService is the domain contract for account operations.
type AccountService interface {
	Create(ctx context.Context, name string) (AccountID, error)
	Delete(ctx context.Context, id AccountID) error
	List(ctx context.Context) ([]Account, error)
}
