package sqlite

import (
	"context"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// Services provides type-safe access to services.
type Services interface {
	Count(ctx context.Context) (int64, error)
	Get(ctx context.Context, id domain.ServiceID) (gen.Service, error)
	SetEnabled(ctx context.Context, id domain.ServiceID, enabled bool) error
}

// servicesImpl implements Services.
type servicesImpl struct {
	read  *gen.Queries
	write *gen.Queries
}

// Count returns the total number of services.
func (s *servicesImpl) Count(ctx context.Context) (int64, error) {
	count, err := s.read.CountServices(ctx)
	if err != nil {
		return 0, WrapSQLiteError(err, "count services")
	}
	return count, nil
}

// Get returns a service by ID.
func (s *servicesImpl) Get(ctx context.Context, id domain.ServiceID) (gen.Service, error) {
	idStr := id.String()
	svc, err := s.read.GetService(ctx, &idStr)
	if err != nil {
		return gen.Service{}, WrapSQLiteError(err, "get service")
	}
	return svc, nil
}

// SetEnabled enables or disables a service.
func (s *servicesImpl) SetEnabled(ctx context.Context, id domain.ServiceID, enabled bool) error {
	var val int64
	if enabled {
		val = 1
	}
	idStr := id.String()
	err := s.write.SetServiceEnabled(ctx, gen.SetServiceEnabledParams{
		Enabled: &val,
		ID:      &idStr,
	})
	if err != nil {
		return WrapSQLiteError(err, "set service enabled")
	}
	return nil
}
