package api

import (
	"context"

	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/pkg/client"
)

// AccountService handles account-related API operations.
type AccountService struct {
	client Client
	logger log.Logger
}

// NewAccountService creates a new account service.
func NewAccountService(client Client, logger log.Logger) *AccountService {
	return &AccountService{
		client: client,
		logger: logger,
	}
}

// Account is the domain model for an account.
type Account struct {
	ID   string
	Name string
}

// List fetches all accounts for an organization.
func (s *AccountService) List(ctx context.Context, organizationID string) ([]Account, error) {
	s.logger.Debug("fetching accounts from API", "organizationID", organizationID)
	resp, err := s.client.ListAccounts(ctx, organizationID)
	if err != nil {
		s.logger.Error("failed to fetch accounts", "error", err, "organizationID", organizationID)
		return nil, err
	}

	// Convert GraphQL response to domain model
	accounts := make([]Account, len(resp.Accounts.Edges))
	for i, edge := range resp.Accounts.Edges {
		accounts[i] = Account{
			ID:   edge.Node.Id,
			Name: edge.Node.Name,
		}
	}

	s.logger.Debug("fetched accounts from API", "count", len(accounts))
	return accounts, nil
}

// Create creates a new account.
func (s *AccountService) Create(ctx context.Context, organizationID, name string) (*Account, error) {
	s.logger.Debug("creating account via API", "organizationID", organizationID, "name", name)
	input := client.CreateAccountInput{
		OrganizationID: organizationID,
		Name:           name,
	}

	resp, err := s.client.CreateAccount(ctx, input)
	if err != nil {
		s.logger.Error("failed to create account", "error", err)
		return nil, err
	}

	account := &Account{
		ID:   resp.CreateAccount.Id,
		Name: resp.CreateAccount.Name,
	}

	s.logger.Debug("created account via API", "id", account.ID, "name", account.Name)
	return account, nil
}
