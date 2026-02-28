package msgs

import "github.com/usetero/cli/internal/domain"

// EnsureRuntime requests runtime initialization for the selected account.
type EnsureRuntime struct {
	Org     domain.Organization
	Account domain.Account
}

// RuntimeReady is emitted when account runtime is initialized.
type RuntimeReady struct {
	Org     domain.Organization
	Account domain.Account
}
