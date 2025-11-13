package api

import (
	"context"

	"github.com/usetero/cli/pkg/client"
)

// Client defines the interface for communicating with the Tero control plane.
// This allows services to be tested without real API calls.
// Concrete implementation: *client.Client (generated GraphQL client)
type Client interface {
	// Organization operations
	ListOrganizations(ctx context.Context) (*client.ListOrganizationsResponse, error)
	CreateOrganizationAndBootstrap(ctx context.Context, input client.CreateOrganizationInput) (*client.CreateOrganizationAndBootstrapResponse, error)

	// Account operations
	ListAccounts(ctx context.Context, organizationID string) (*client.ListAccountsResponse, error)
	CreateAccount(ctx context.Context, input client.CreateAccountInput) (*client.CreateAccountResponse, error)
	GetAccount(ctx context.Context, accountID string) (*client.GetAccountResponse, error)

	// Datadog operations
	ValidateDatadogApiKey(ctx context.Context, input client.ValidateDatadogApiKeyInput) (*client.ValidateDatadogApiKeyResponse, error)
	CreateDatadogAccountWithCredentials(ctx context.Context, input client.CreateDatadogAccountWithCredentialsInput) (*client.CreateDatadogAccountWithCredentialsResponse, error)
	GetDatadogAccountServiceDiscoveryProgress(ctx context.Context, id string) (*client.GetDatadogAccountServiceDiscoveryProgressResponse, error)
	GetDatadogAccountLogDiscoveryProgress(ctx context.Context, id string) (*client.GetDatadogAccountLogDiscoveryProgressResponse, error)
}
