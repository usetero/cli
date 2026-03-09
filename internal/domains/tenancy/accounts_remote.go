package tenancy

import (
	"context"
	"fmt"

	controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"
)

type remoteAccountClient interface {
	CreateAccount(ctx context.Context, organizationID controlplane.OrganizationID, name string) (controlplane.Account, error)
	DeleteAccount(ctx context.Context, accountID controlplane.AccountID) error
	ListAccounts(ctx context.Context, organizationID controlplane.OrganizationID) ([]controlplane.Account, error)
}

// RemoteAccountService uses control-plane API for account operations.
type RemoteAccountService struct {
	client         remoteAccountClient
	organizationID OrganizationID
}

func NewRemoteAccountService(client remoteAccountClient, organizationID OrganizationID) *RemoteAccountService {
	if client == nil {
		panic("tenancy remote account service requires client")
	}
	if organizationID == "" {
		panic("tenancy remote account service requires organization id")
	}
	return &RemoteAccountService{client: client, organizationID: organizationID}
}

func (s *RemoteAccountService) Create(ctx context.Context, create AccountCreate) (AccountID, error) {
	validated, err := create.Validate()
	if err != nil {
		return "", err
	}

	account, err := s.client.CreateAccount(ctx, toControlPlaneOrganizationID(s.organizationID), validated.Name)
	if err != nil {
		return "", err
	}
	return fromControlPlaneAccount(account).ID, nil
}

func (s *RemoteAccountService) Delete(ctx context.Context, id AccountID) error {
	if id == "" {
		return fmt.Errorf("account id is required")
	}
	return s.client.DeleteAccount(ctx, toControlPlaneAccountID(id))
}

func (s *RemoteAccountService) List(ctx context.Context) ([]Account, error) {
	rows, err := s.client.ListAccounts(ctx, toControlPlaneOrganizationID(s.organizationID))
	if err != nil {
		return nil, err
	}

	out := make([]Account, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromControlPlaneAccount(row))
	}
	return out, nil
}
