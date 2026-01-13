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

// GetDatadogAccountStatus gets the discovery status for a Datadog account
func (c *Client) GetDatadogAccountStatus(ctx context.Context, id string) (*GetDatadogAccountStatusResponse, error) {
	return GetDatadogAccountStatus(ctx, c.gql, id)
}
