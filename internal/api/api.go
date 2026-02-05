package api

import (
	"github.com/usetero/cli/internal/log"
)

// API bundles all control plane API services.
// It provides a single entry point for all API-related operations.
type API struct {
	Organizations   *OrganizationService
	Accounts        *AccountService
	DatadogAccounts *DatadogAccountService
	Services        *ServiceService
}

// New creates a new API with all services initialized.
// Requires an authenticated API client.
func New(client Client, scope log.Scope) *API {
	scope = scope.Child("api")
	return &API{
		Organizations:   NewOrganizationService(client, scope),
		Accounts:        NewAccountService(client, scope),
		DatadogAccounts: NewDatadogAccountService(client, scope),
		Services:        NewServiceService(client, scope),
	}
}
