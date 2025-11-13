package client

import "context"

// ListServices returns all services
func (c *Client) ListServices(ctx context.Context) (*ListServicesResponse, error) {
	return ListServices(ctx, c.gql)
}

// GetService retrieves a specific service by ID
func (c *Client) GetService(ctx context.Context, id string) (*GetServiceResponse, error) {
	return GetService(ctx, c.gql, id)
}

// GetServiceByName retrieves a specific service by name
func (c *Client) GetServiceByName(ctx context.Context, name string) (*GetServiceByNameResponse, error) {
	return GetServiceByName(ctx, c.gql, name)
}

// EnableService enables a service for analysis
func (c *Client) EnableService(ctx context.Context, serviceId string) (*EnableServiceResponse, error) {
	return EnableService(ctx, c.gql, serviceId)
}
