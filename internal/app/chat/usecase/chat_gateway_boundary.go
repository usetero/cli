package usecase

import (
	"context"

	chatboundary "github.com/usetero/cli/internal/boundary/chat"
	corechat "github.com/usetero/cli/internal/core/chat"
)

// ChatBoundaryGateway adapts chatboundary.Client to the use-case ChatGateway.
type ChatBoundaryGateway struct {
	client chatboundary.Client
}

func NewChatBoundaryGateway(client chatboundary.Client) *ChatBoundaryGateway {
	return &ChatBoundaryGateway{client: client}
}

func (g *ChatBoundaryGateway) StreamSnapshots(ctx context.Context, req StreamRequest, onSnapshot func(corechat.StreamSnapshot)) (*corechat.StreamResult, error) {
	if g == nil || g.client == nil {
		return nil, nil
	}
	wireReq := chatboundary.Request{
		ConversationID:  req.ConversationID.String(),
		Messages:        req.Messages,
		ContextEntities: req.ContextEntities,
	}
	return g.client.StreamSnapshots(ctx, wireReq, onSnapshot)
}
