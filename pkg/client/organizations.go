package client

import "context"

// ListOrganizations returns all organizations for the authenticated user
func (c *Client) ListOrganizations(ctx context.Context) (*ListOrganizationsResponse, error) {
	return ListOrganizations(ctx, c.gql)
}

// CreateOrganizationAndBootstrap creates a new organization with default setup
func (c *Client) CreateOrganizationAndBootstrap(ctx context.Context, input CreateOrganizationInput) (*CreateOrganizationAndBootstrapResponse, error) {
	return CreateOrganizationAndBootstrap(ctx, c.gql, input)
}
