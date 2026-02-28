package msgs

import "github.com/usetero/cli/internal/domain"

// AccountSelected is emitted when user selects an account.
type AccountSelected struct {
	Org     domain.Organization
	Account domain.Account
}

// NoAccounts is emitted when no accounts exist for the org.
type NoAccounts struct {
	Org domain.Organization
}

// AccountCreated is emitted when a new account is created.
type AccountCreated struct {
	Org     domain.Organization
	Account domain.Account
}
