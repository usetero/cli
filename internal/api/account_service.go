package api

import (
	"context"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/api/gen"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
)

// Accounts provides access to accounts.
type Accounts interface {
	List(ctx context.Context, organizationID domain.OrganizationID) ([]domain.Account, error)
	Get(ctx context.Context, accountID domain.AccountID) (*domain.Account, error)
	Create(ctx context.Context, id uuid.UUID, organizationID domain.OrganizationID, name string) (*domain.Account, error)
}

// AccountService handles account-related API operations.
type AccountService struct {
	client Client
	logger log.Logger
}

// Ensure AccountService implements Accounts.
var _ Accounts = (*AccountService)(nil)

// NewAccountService creates a new account service.
func NewAccountService(client Client, logger log.Logger) *AccountService {
	return &AccountService{
		client: client,
		logger: logger,
	}
}

// List fetches all accounts for an organization.
func (s *AccountService) List(ctx context.Context, organizationID domain.OrganizationID) ([]domain.Account, error) {
	s.logger.Debug("fetching accounts from API", "organizationID", organizationID)
	resp, err := s.client.ListAccounts(ctx, organizationID.String())
	if err != nil {
		s.logger.Error("failed to fetch accounts", "error", err, "organizationID", organizationID)
		return nil, err
	}

	// Convert GraphQL response to domain model
	accounts := make([]domain.Account, len(resp.Accounts.Edges))
	for i, edge := range resp.Accounts.Edges {
		accounts[i] = domain.Account{
			ID:   domain.AccountID(edge.Node.Id),
			Name: edge.Node.Name,
		}
	}

	s.logger.Debug("fetched accounts from API", "count", len(accounts))
	return accounts, nil
}

// Get fetches a single account by ID. Returns nil if not found.
func (s *AccountService) Get(ctx context.Context, accountID domain.AccountID) (*domain.Account, error) {
	s.logger.Debug("fetching account from API", "accountID", accountID)
	resp, err := s.client.GetAccount(ctx, accountID.String())
	if err != nil {
		s.logger.Error("failed to fetch account", "error", err, "accountID", accountID)
		return nil, err
	}

	s.logger.Debug("GetAccount response", "edges", len(resp.Accounts.Edges))
	if len(resp.Accounts.Edges) == 0 {
		s.logger.Debug("account not found", "accountID", accountID)
		return nil, nil
	}

	node := resp.Accounts.Edges[0].Node
	return &domain.Account{
		ID:   domain.AccountID(node.Id),
		Name: node.DatadogAccount.Name,
	}, nil
}

// Create creates a new account with the given client-provided ID.
func (s *AccountService) Create(ctx context.Context, id uuid.UUID, organizationID domain.OrganizationID, name string) (*domain.Account, error) {
	s.logger.Debug("creating account via API", "id", id.String(), "organizationID", organizationID, "name", name)
	input := gen.CreateAccountInput{
		Id:             ptr(id.String()),
		OrganizationID: organizationID.String(),
		Name:           name,
	}

	resp, err := s.client.CreateAccount(ctx, input)
	if err != nil {
		s.logger.Error("failed to create account", "error", err)
		return nil, err
	}

	account := &domain.Account{
		ID:   domain.AccountID(resp.CreateAccount.Id),
		Name: resp.CreateAccount.Name,
	}

	s.logger.Debug("created account via API", "id", account.ID, "name", account.Name)
	return account, nil
}
