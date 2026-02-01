package api

import (
	"context"

	"github.com/usetero/cli/pkg/client"
)

// Client defines the interface for communicating with the Tero control plane.
// This allows services to be tested without real API calls.
// Concrete implementation: *client.Client (generated GraphQL client)
type Client interface {
	// SetAccessToken updates the access token used for authentication.
	SetAccessToken(token string)

	// SetAccountID sets the account ID header for scoped requests.
	SetAccountID(accountID string)

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
	GetDatadogAccountStatus(ctx context.Context, id string) (*client.GetDatadogAccountStatusResponse, error)

	// Conversation operations
	CreateConversation(ctx context.Context, input client.CreateConversationInput) (*client.CreateConversationResponse, error)
}
