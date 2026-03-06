package session

import (
	"fmt"
)

func validateStartAccountID(accountID string) error {
	if accountID == "" {
		return fmt.Errorf("account id is required")
	}
	return nil
}
