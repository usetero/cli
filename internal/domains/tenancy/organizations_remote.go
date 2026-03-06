package tenancy

import (
	"context"

	controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"
)

type remoteOrganizationClient interface {
	ListOrganizations(ctx context.Context) ([]controlplane.Organization, error)
	CreateOrganizationAndBootstrap(ctx context.Context, name string) (controlplane.OrganizationBootstrap, error)
}

// RemoteOrganizationService uses control-plane API for organizations.
type RemoteOrganizationService struct {
	client remoteOrganizationClient
}

func NewRemoteOrganizationService(client remoteOrganizationClient) *RemoteOrganizationService {
	if client == nil {
		panic("tenancy remote organization service requires client")
	}
	return &RemoteOrganizationService{client: client}
}

func (s *RemoteOrganizationService) List(ctx context.Context) ([]Organization, error) {
	orgs, err := s.client.ListOrganizations(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]Organization, 0, len(orgs))
	for _, org := range orgs {
		out = append(out, fromControlPlaneOrganization(org))
	}
	return out, nil
}

func (s *RemoteOrganizationService) Create(ctx context.Context, create OrganizationCreate) (OrganizationBootstrap, error) {
	validated, err := create.Validate()
	if err != nil {
		return OrganizationBootstrap{}, err
	}

	bootstrap, err := s.client.CreateOrganizationAndBootstrap(ctx, validated.Name)
	if err != nil {
		return OrganizationBootstrap{}, err
	}

	return fromControlPlaneBootstrap(bootstrap), nil
}
