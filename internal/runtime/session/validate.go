package session

import (
	"fmt"

	"github.com/usetero/cli/internal/domains/tenancy"
)

func (s *Service) validateStart(accountID tenancy.AccountID) error {
	if s == nil {
		return fmt.Errorf("session service is nil")
	}
	if accountID == "" {
		return fmt.Errorf("account id is required")
	}
	if s.storage == nil {
		return fmt.Errorf("storage dependency is required")
	}
	if s.newSyncer == nil {
		return fmt.Errorf("syncer factory is required")
	}
	if s.newUploader == nil {
		return fmt.Errorf("uploader factory is required")
	}
	if s.openDB == nil {
		return fmt.Errorf("database open function is required")
	}
	return nil
}
