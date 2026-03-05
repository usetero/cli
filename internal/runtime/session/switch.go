package session

import (
	"context"

	"github.com/usetero/cli/internal/domains/tenancy"
)

// Switch moves lifecycle to a new account (Stop + Start).
func (s *Service) Switch(ctx context.Context, accountID tenancy.AccountID) error {
	if err := s.Stop(); err != nil {
		return err
	}
	return s.Start(ctx, accountID)
}
