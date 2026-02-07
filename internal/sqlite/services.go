package sqlite

import (
	"context"

	"github.com/usetero/cli/internal/sqlite/gen"
)

// Services provides type-safe access to services.
type Services interface {
	Count(ctx context.Context) (int64, error)
}

// servicesImpl implements Services.
type servicesImpl struct {
	queries *gen.Queries
}

// Count returns the total number of services.
func (s *servicesImpl) Count(ctx context.Context) (int64, error) {
	count, err := s.queries.CountServices(ctx)
	if err != nil {
		return 0, WrapSQLiteError(err, "count services")
	}
	return count, nil
}
