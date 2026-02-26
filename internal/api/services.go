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
	scope           log.Scope
	Organizations   Organizations
	Accounts        Accounts
	Workspaces      Workspaces
	DatadogAccounts DatadogAccounts
	Conversations   Conversations
	Messages        Messages
	Services        Services
	Policies        Policies
}

// NewServices creates APIServices with an internally-managed client.
// This is the preferred constructor for production use.
func NewServices(endpoint string, authService auth.Auth, scope log.Scope) APIServices {
	c := NewClient(endpoint, authService)
	return newAPIServices(c, scope)
}

// NewAPIServices creates all API services from the given client.
// Use this constructor when you need to inject a mock client for testing.
func NewAPIServices(client Client, scope log.Scope) APIServices {
	return newAPIServices(client, scope)
}

func newAPIServices(client Client, scope log.Scope) APIServices {
	return newAPIServicesWithScope(client, scope.Child("api"))
}

func newAPIServicesWithScope(client Client, scope log.Scope) APIServices {
	return APIServices{
		client:          client,
		scope:           scope,
		Organizations:   NewOrganizationService(client, scope),
		Accounts:        NewAccountService(client, scope),
		Workspaces:      NewWorkspaceService(client, scope),
		DatadogAccounts: NewDatadogAccountService(client, scope),
		Conversations:   NewConversationService(client, scope),
		Messages:        NewMessageService(client, scope),
		Services:        NewServiceService(client, scope),
		Policies:        NewPolicyService(client, scope),
	}
}

// SetAccountID sets the account ID header for scoped requests.
func (s APIServices) SetAccountID(accountID domain.AccountID) {
	s.client.SetAccountID(accountID)
}

// WithAccountID returns a new APIServices value scoped to accountID.
func (s APIServices) WithAccountID(accountID domain.AccountID) APIServices {
	return newAPIServicesWithScope(s.client.WithAccountID(accountID), s.scope)
}

// RawQuery executes an arbitrary GraphQL query (for debugging).
func (s APIServices) RawQuery(ctx context.Context, query string, variables map[string]interface{}) (map[string]interface{}, error) {
	return s.client.RawQuery(ctx, query, variables)
}
