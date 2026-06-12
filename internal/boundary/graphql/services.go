package graphql

import (
	"context"

	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
)

// ServiceSet aggregates all API services for easy dependency injection.
// This is the primary public interface for interacting with the Tero API.
type ServiceSet struct {
	client          Client
	scope           log.Scope
	Organizations   Organizations
	Accounts        Accounts
	Workspaces      Workspaces
	DatadogAccounts DatadogAccounts
	Services        Services
	Policies        Policies
	Issues          Issues
	Checks          Checks
	EdgeInstances   EdgeInstances
}

// NewServiceSet creates ServiceSet with an internally-managed client.
// This is the preferred constructor for production use.
func NewServiceSet(endpoint string, authService auth.Auth, scope log.Scope) ServiceSet {
	c := NewClient(endpoint, authService)
	return newServiceSet(c, scope)
}

// NewServiceSetFromClient creates all API services from the given client.
// Use this constructor when you need to inject a mock client for testing.
func NewServiceSetFromClient(client Client, scope log.Scope) ServiceSet {
	return newServiceSet(client, scope)
}

func newServiceSet(client Client, scope log.Scope) ServiceSet {
	return newServiceSetWithScope(client, scope.Child("api"))
}

func newServiceSetWithScope(client Client, scope log.Scope) ServiceSet {
	return ServiceSet{
		client:          client,
		scope:           scope,
		Organizations:   NewOrganizationService(client, scope),
		Accounts:        NewAccountService(client, scope),
		Workspaces:      NewWorkspaceService(client, scope),
		DatadogAccounts: NewDatadogAccountService(client, scope),
		Services:        NewServiceService(client, scope),
		Policies:        NewPolicyService(client, scope),
		Issues:          NewIssueService(client, scope),
		Checks:          NewCheckService(client, scope),
		EdgeInstances:   NewEdgeInstanceService(client, scope),
	}
}

// SetAccountID sets the account ID header for scoped requests.
func (s ServiceSet) SetAccountID(accountID domain.AccountID) {
	s.client.SetAccountID(accountID)
}

// WithAccountID returns a new ServiceSet value scoped to accountID.
func (s ServiceSet) WithAccountID(accountID domain.AccountID) ServiceSet {
	return newServiceSetWithScope(s.client.WithAccountID(accountID), s.scope)
}

// RawQuery executes an arbitrary GraphQL query (for debugging).
func (s ServiceSet) RawQuery(ctx context.Context, query string, variables map[string]interface{}) (map[string]interface{}, error) {
	return s.client.RawQuery(ctx, query, variables)
}
