package client

import "context"

// CreateConversation creates a new conversation
func (c *Client) CreateConversation(ctx context.Context, input CreateConversationInput) (*CreateConversationResponse, error) {
	return CreateConversation(ctx, c.gql, input)
}
