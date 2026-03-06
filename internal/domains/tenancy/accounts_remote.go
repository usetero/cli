package tenancy

import (
	"context"
	"fmt"

	controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"
)

type remoteAccountClient interface {
	CreateAccount(ctx context.Context, organizationID controlplane.OrganizationID, name string) (controlplane.Account, error)
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

func (s *RemoteAccountService) Delete(_ context.Context, _ AccountID) error {
	return fmt.Errorf("tenancy remote account delete is not implemented")
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
