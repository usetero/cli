package tenancy

import (
	"context"
	"fmt"

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
	return &RemoteOrganizationService{client: client}
}

func (s *RemoteOrganizationService) List(ctx context.Context) ([]Organization, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("tenancy remote organization service is not initialized")
	}
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

func (s *RemoteOrganizationService) Create(ctx context.Context, name string) (OrganizationBootstrap, error) {
	if s == nil || s.client == nil {
		return OrganizationBootstrap{}, fmt.Errorf("tenancy remote organization service is not initialized")
	}
	if name == "" {
		return OrganizationBootstrap{}, fmt.Errorf("organization name is required")
	}

	bootstrap, err := s.client.CreateOrganizationAndBootstrap(ctx, name)
	if err != nil {
		return OrganizationBootstrap{}, err
	}

	return fromControlPlaneBootstrap(bootstrap), nil
}
