package client

import "context"

// ListAccounts returns all accounts for a given organization
func (c *Client) ListAccounts(ctx context.Context, organizationID string) (*ListAccountsResponse, error) {
	return ListAccounts(ctx, c.gql, organizationID)
}

// CreateAccount creates a new account within an organization
func (c *Client) CreateAccount(ctx context.Context, input CreateAccountInput) (*CreateAccountResponse, error) {
	return CreateAccount(ctx, c.gql, input)
}

// GetAccount retrieves a specific account by ID
func (c *Client) GetAccount(ctx context.Context, id string) (*GetAccountResponse, error) {
	return GetAccount(ctx, c.gql, id)
}
