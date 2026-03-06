package syncertest

import (
	"context"

	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
)

// MockReadinessService is a functional mock for syncer.ReadinessService.
type MockReadinessService struct {
	ReadyFn func(ctx context.Context) (bool, error)
}

var _ pssyncer.ReadinessService = (*MockReadinessService)(nil)

func (m MockReadinessService) Ready(ctx context.Context) (bool, error) {
	if m.ReadyFn == nil {
		return false, nil
	}
	return m.ReadyFn(ctx)
}
