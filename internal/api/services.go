package api

import (
	"context"

	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
)

// APIServices aggregates all API services for easy dependency injection.
// This is the primary public interface for interacting with the Tero API.
type APIServices struct {
	client          Client
	Organizations   Organizations
	Accounts        Accounts
	Workspaces      Workspaces
	DatadogAccounts DatadogAccounts
	Conversations   Conversations
	Messages        Messages
}

// NewServices creates APIServices with an internally-managed client.
// This is the preferred constructor for production use.
func NewServices(endpoint string, authService auth.Auth, logger log.Logger) APIServices {
	c := NewClient(endpoint, authService)
	return newAPIServices(c, logger)
}

// NewAPIServices creates all API services from the given client.
// Use this constructor when you need to inject a mock client for testing.
func NewAPIServices(client Client, logger log.Logger) APIServices {
	return newAPIServices(client, logger)
}

func newAPIServices(client Client, logger log.Logger) APIServices {
	return APIServices{
		client:          client,
		Organizations:   NewOrganizationService(client, logger),
		Accounts:        NewAccountService(client, logger),
		Workspaces:      NewWorkspaceService(client, logger),
		DatadogAccounts: NewDatadogAccountService(client, logger),
		Conversations:   NewConversationService(client, logger),
		Messages:        NewMessageService(client, logger),
	}
}

// SetAccountID sets the account ID header for scoped requests.
func (s APIServices) SetAccountID(accountID domain.AccountID) {
	s.client.SetAccountID(accountID)
}

// RawQuery executes an arbitrary GraphQL query (for debugging).
func (s APIServices) RawQuery(ctx context.Context, query string, variables map[string]interface{}) (map[string]interface{}, error) {
	return s.client.RawQuery(ctx, query, variables)
}
