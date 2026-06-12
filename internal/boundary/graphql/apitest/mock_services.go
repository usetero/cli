package apitest

import graphql "github.com/usetero/cli/internal/boundary/graphql"

// NewMockServiceSet creates a ServiceSet with mock implementations.
// Pass nil for any service you don't need to mock.
func NewMockServiceSet(
	organizations *MockOrganizations,
	accounts *MockAccounts,
	datadogAccounts *MockDatadogAccounts,
) graphql.ServiceSet {
	services := graphql.ServiceSet{}

	if organizations != nil {
		services.Organizations = organizations
	}
	if accounts != nil {
		services.Accounts = accounts
	}
	if datadogAccounts != nil {
		services.DatadogAccounts = datadogAccounts
	}

	return services
}
