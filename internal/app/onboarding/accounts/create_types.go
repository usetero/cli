package accounts

import "github.com/usetero/cli/internal/domain"

// accountCreatedMsg is sent when account creation completes.
type accountCreatedMsg struct {
	account domain.Account
	err     error
}
