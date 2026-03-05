package tenancy

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	accountsdb "github.com/usetero/cli/internal/domains/tenancy/db/accountsgen"
)

// LocalAccountService uses SQLite/sqlc for account CRUD.
type LocalAccountService struct {
	q *accountsdb.Queries
}

func NewLocalAccountService(db *sql.DB) *LocalAccountService {
	return &LocalAccountService{q: accountsdb.New(db)}
}

func (s *LocalAccountService) Create(ctx context.Context, name string) (AccountID, error) {
	if s == nil || s.q == nil {
		return "", fmt.Errorf("tenancy local account service is not initialized")
	}
	if name == "" {
		return "", fmt.Errorf("account name is required")
	}

	id := AccountID(uuid.NewString())
	err := s.q.Create(ctx, accountsdb.CreateParams{
		ID:        toAccountsDBAccountID(id),
		Name:      name,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *LocalAccountService) Delete(ctx context.Context, id AccountID) error {
	if s == nil || s.q == nil {
		return fmt.Errorf("tenancy local account service is not initialized")
	}
	if id == "" {
		return fmt.Errorf("account id is required")
	}
	return s.q.Delete(ctx, toAccountsDBAccountID(id))
}

func (s *LocalAccountService) List(ctx context.Context) ([]Account, error) {
	if s == nil || s.q == nil {
		return nil, fmt.Errorf("tenancy local account service is not initialized")
	}
	rows, err := s.q.List(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]Account, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromAccountsDBAccount(row))
	}
	return out, nil
}
