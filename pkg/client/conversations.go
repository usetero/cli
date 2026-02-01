package client

import "context"

// CreateConversation creates a new conversation
func (c *Client) CreateConversation(ctx context.Context, input CreateConversationInput) (*CreateConversationResponse, error) {
	return CreateConversation(ctx, c.gql, input)
}

// UpdateConversation updates a conversation
func (c *Client) UpdateConversation(ctx context.Context, id string, input UpdateConversationInput) (*UpdateConversationResponse, error) {
	return UpdateConversation(ctx, c.gql, id, input)
}

// DeleteConversation deletes a conversation
func (c *Client) DeleteConversation(ctx context.Context, id string) (*DeleteConversationResponse, error) {
	return DeleteConversation(ctx, c.gql, id)
}
