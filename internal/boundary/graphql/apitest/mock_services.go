package apitest

import graphql "github.com/usetero/cli/internal/boundary/graphql"

// NewMockAPIServices creates an APIServices with mock implementations.
// Pass nil for any service you don't need to mock.
func NewMockAPIServices(
	organizations *MockOrganizations,
	accounts *MockAccounts,
	workspaces *MockWorkspaces,
	datadogAccounts *MockDatadogAccounts,
) graphql.APIServices {
	services := graphql.APIServices{}

	if organizations != nil {
		services.Organizations = organizations
	}
	if accounts != nil {
		services.Accounts = accounts
	}
	if workspaces != nil {
		services.Workspaces = workspaces
	}
	if datadogAccounts != nil {
		services.DatadogAccounts = datadogAccounts
	}

	return services
}
