package change

import (
	"context"
	"database/sql"
	"fmt"

	findingsdb "github.com/usetero/cli/internal/domains/change/db/findingsgen"
)

// LocalFindingService uses SQLite/sqlc for local finding reads.
type LocalFindingService struct {
	q *findingsdb.Queries
}

func NewLocalFindingService(db *sql.DB) *LocalFindingService {
	if db == nil {
		panic("change local service requires db")
	}
	return &LocalFindingService{q: findingsdb.New(db)}
}

func (s *LocalFindingService) List(ctx context.Context) ([]Finding, error) {
	rows, err := s.q.ListFindings(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Finding, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromListRow(row))
	}
	return out, nil
}

func (s *LocalFindingService) ListByDomain(ctx context.Context, domain Domain) ([]Finding, error) {
	if domain == "" {
		return nil, fmt.Errorf("finding domain is required")
	}
	rows, err := s.q.ListFindingsByDomain(ctx, ptrString(string(domain)))
	if err != nil {
		return nil, err
	}
	out := make([]Finding, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromDomainRow(row))
	}
	return out, nil
}

func (s *LocalFindingService) ListByService(ctx context.Context, serviceID ServiceID) ([]Finding, error) {
	if serviceID == "" {
		return nil, fmt.Errorf("service id is required")
	}
	rows, err := s.q.ListFindingsByService(ctx, ptrString(string(serviceID)))
	if err != nil {
		return nil, err
	}
	out := make([]Finding, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromServiceRow(row))
	}
	return out, nil
}
