package usecase

import (
	"context"

	chatclient "github.com/usetero/cli/internal/boundary/chat"
	corechat "github.com/usetero/cli/internal/core/chat"
)

// ChatClientGateway adapts chatclient.Client to the use-case ChatGateway.
type ChatClientGateway struct {
	client chatclient.Client
}

func NewChatClientGateway(client chatclient.Client) *ChatClientGateway {
	return &ChatClientGateway{client: client}
}

func (g *ChatClientGateway) StreamSnapshots(ctx context.Context, req StreamRequest, onSnapshot func(corechat.StreamSnapshot)) (*corechat.StreamResult, error) {
	if g == nil || g.client == nil {
		return nil, nil
	}
	wireReq := chatclient.Request{
		ConversationID:  req.ConversationID.String(),
		Messages:        req.Messages,
		ContextEntities: req.ContextEntities,
	}
	return g.client.StreamSnapshots(ctx, wireReq, onSnapshot)
}
