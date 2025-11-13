package client

import "context"

// CreateDatadogAccountWithCredentials creates a new Datadog account with API and App keys
func (c *Client) CreateDatadogAccountWithCredentials(ctx context.Context, input CreateDatadogAccountWithCredentialsInput) (*CreateDatadogAccountWithCredentialsResponse, error) {
	return CreateDatadogAccountWithCredentials(ctx, c.gql, input)
}

// ValidateDatadogApiKey validates a Datadog API key
func (c *Client) ValidateDatadogApiKey(ctx context.Context, input ValidateDatadogApiKeyInput) (*ValidateDatadogApiKeyResponse, error) {
	return ValidateDatadogApiKey(ctx, c.gql, input)
}

// GetDatadogAccountServiceDiscoveryProgress gets the service discovery progress for a Datadog account
func (c *Client) GetDatadogAccountServiceDiscoveryProgress(ctx context.Context, id string) (*GetDatadogAccountServiceDiscoveryProgressResponse, error) {
	return GetDatadogAccountServiceDiscoveryProgress(ctx, c.gql, id)
}

// GetDatadogAccountLogDiscoveryProgress gets the log event discovery progress for a Datadog account
func (c *Client) GetDatadogAccountLogDiscoveryProgress(ctx context.Context, id string) (*GetDatadogAccountLogDiscoveryProgressResponse, error) {
	return GetDatadogAccountLogDiscoveryProgress(ctx, c.gql, id)
}
