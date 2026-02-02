package client

import "context"

// ListWorkspaces returns all workspaces for a given account
func (c *Client) ListWorkspaces(ctx context.Context, accountID string) (*ListWorkspacesResponse, error) {
	return ListWorkspaces(ctx, c.gql, accountID)
}
