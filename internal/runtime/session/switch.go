package session

import (
	"context"

	"github.com/usetero/cli/internal/domains/tenancy"
)

// Switch moves lifecycle to a new account (Stop + Start).
func (s *Service) Switch(ctx context.Context, accountID tenancy.AccountID) error {
	if err := validateStartAccountID(string(accountID)); err != nil {
		return err
	}
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if err := s.stopLocked(); err != nil {
		return err
	}
	return s.startLocked(ctx, accountID)
}
