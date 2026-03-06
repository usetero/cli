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
	if db == nil {
		panic("tenancy local account service requires db")
	}
	return &LocalAccountService{q: accountsdb.New(db)}
}

func (s *LocalAccountService) Create(ctx context.Context, create AccountCreate) (AccountID, error) {
	validated, err := create.Validate()
	if err != nil {
		return "", err
	}

	id := AccountID(uuid.NewString())
	err = s.q.Create(ctx, accountsdb.CreateParams{
		ID:        toAccountsDBAccountID(id),
		Name:      validated.Name,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *LocalAccountService) Delete(ctx context.Context, id AccountID) error {
	if id == "" {
		return fmt.Errorf("account id is required")
	}
	return s.q.Delete(ctx, toAccountsDBAccountID(id))
}

func (s *LocalAccountService) List(ctx context.Context) ([]Account, error) {
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
