package chattest

import (
	"context"

	domainchat "github.com/usetero/cli/internal/domains/chat"
)

type MessageService struct {
	CreateUserMessageFn  func(ctx context.Context, create domainchat.UserMessageCreate) (domainchat.MessageID, error)
	DeleteFn             func(ctx context.Context, id domainchat.MessageID) error
	ListByConversationFn func(ctx context.Context, conversationID domainchat.ConversationID) ([]domainchat.Message, error)
}

var _ domainchat.MessageService = (*MessageService)(nil)

func NewMessageService() *MessageService {
	return &MessageService{}
}

func (s *MessageService) CreateUserMessage(ctx context.Context, create domainchat.UserMessageCreate) (domainchat.MessageID, error) {
	if s.CreateUserMessageFn == nil {
		return "", nil
	}
	return s.CreateUserMessageFn(ctx, create)
}

func (s *MessageService) Delete(ctx context.Context, id domainchat.MessageID) error {
	if s.DeleteFn == nil {
		return nil
	}
	return s.DeleteFn(ctx, id)
}

func (s *MessageService) ListByConversation(ctx context.Context, conversationID domainchat.ConversationID) ([]domainchat.Message, error) {
	if s.ListByConversationFn == nil {
		return nil, nil
	}
	return s.ListByConversationFn(ctx, conversationID)
}
