package api

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/infrastructure/controlplane/api/gen"
)

// ListOrganizations fetches organizations for the current user.
func (c *BootstrapClient) ListOrganizations(ctx context.Context) ([]Organization, error) {
	client, err := c.base.gql(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gen.ListOrganizations(ctx, client)
	if err != nil {
		return nil, err
	}

	edges := resp.Organizations.Edges
	out := make([]Organization, 0, len(edges))
	for _, edge := range edges {
		if edge == nil {
			continue
		}
		node := edge.Node
		out = append(out, Organization{
			ID:                   OrganizationID(node.Id),
			Name:                 node.Name,
			WorkosOrganizationID: node.WorkosOrganizationID,
		})
	}
	return out, nil
}

// ListAccounts fetches accounts for the given organization.
func (c *BootstrapClient) ListAccounts(ctx context.Context, organizationID OrganizationID) ([]Account, error) {
	if organizationID == "" {
		return nil, fmt.Errorf("organization id is required")
	}

	client, err := c.base.gql(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gen.ListAccounts(ctx, client, organizationID.String())
	if err != nil {
		return nil, err
	}

	edges := resp.Accounts.Edges
	out := make([]Account, 0, len(edges))
	for _, edge := range edges {
		if edge == nil || edge.Node == nil {
			continue
		}
		node := edge.Node
		out = append(out, Account{
			ID:        AccountID(node.Id),
			Name:      node.Name,
			CreatedAt: node.CreatedAt,
		})
	}
	return out, nil
}

// CreateAccount creates an account in the given organization.
func (c *BootstrapClient) CreateAccount(ctx context.Context, organizationID OrganizationID, name string) (Account, error) {
	if organizationID == "" {
		return Account{}, fmt.Errorf("organization id is required")
	}
	if name == "" {
		return Account{}, fmt.Errorf("account name is required")
	}

	client, err := c.base.gql(ctx)
	if err != nil {
		return Account{}, err
	}

	resp, err := gen.CreateAccount(ctx, client, gen.CreateAccountInput{
		Name:           name,
		OrganizationID: organizationID.String(),
	})
	if err != nil {
		return Account{}, err
	}

	return Account{
		ID:        AccountID(resp.CreateAccount.Id),
		Name:      resp.CreateAccount.Name,
		CreatedAt: resp.CreateAccount.CreatedAt,
	}, nil
}

// DeleteAccount deletes an account by ID.
func (c *BootstrapClient) DeleteAccount(ctx context.Context, accountID AccountID) error {
	if accountID == "" {
		return fmt.Errorf("account id is required")
	}

	client, err := c.base.gql(ctx)
	if err != nil {
		return err
	}

	resp, err := gen.DeleteAccount(ctx, client, accountID.String(), "DELETE")
	if err != nil {
		return err
	}
	if !resp.DeleteAccount {
		return fmt.Errorf("delete account returned false")
	}
	return nil
}

// CreateOrganizationAndBootstrap creates org/account/workspace in one mutation.
func (c *BootstrapClient) CreateOrganizationAndBootstrap(ctx context.Context, name string) (OrganizationBootstrap, error) {
	if name == "" {
		return OrganizationBootstrap{}, fmt.Errorf("organization name is required")
	}

	client, err := c.base.gql(ctx)
	if err != nil {
		return OrganizationBootstrap{}, err
	}

	resp, err := gen.CreateOrganizationAndBootstrap(ctx, client, gen.CreateOrganizationInput{Name: name})
	if err != nil {
		return OrganizationBootstrap{}, err
	}

	result := resp.CreateOrganizationAndBootstrap
	return OrganizationBootstrap{
		Organization: Organization{
			ID:                   OrganizationID(result.Organization.Id),
			Name:                 result.Organization.Name,
			WorkosOrganizationID: result.Organization.WorkosOrganizationID,
		},
		Account: Account{
			ID:   AccountID(result.Account.Id),
			Name: result.Account.Name,
		},
		Workspace: Workspace{
			ID:   WorkspaceID(result.Workspace.Id),
			Name: result.Workspace.Name,
		},
	}, nil
}

// DeleteOrganization deletes an organization by ID.
func (c *BootstrapClient) DeleteOrganization(ctx context.Context, organizationID OrganizationID) error {
	if organizationID == "" {
		return fmt.Errorf("organization id is required")
	}

	client, err := c.base.gql(ctx)
	if err != nil {
		return err
	}

	resp, err := gen.DeleteOrganization(ctx, client, organizationID.String(), "DELETE")
	if err != nil {
		return err
	}
	if !resp.DeleteOrganization {
		return fmt.Errorf("delete organization returned false")
	}
	return nil
}
