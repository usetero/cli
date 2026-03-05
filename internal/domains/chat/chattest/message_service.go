package chattest

import (
	"context"

	domainchat "github.com/usetero/cli/internal/domains/chat"
)

type MessageService struct {
	CreateUserMessageFn  func(ctx context.Context, conversationID domainchat.ConversationID, content string) (domainchat.MessageID, error)
	DeleteFn             func(ctx context.Context, id domainchat.MessageID) error
	ListByConversationFn func(ctx context.Context, conversationID domainchat.ConversationID) ([]domainchat.Message, error)
}

func (s *MessageService) CreateUserMessage(ctx context.Context, conversationID domainchat.ConversationID, content string) (domainchat.MessageID, error) {
	if s.CreateUserMessageFn == nil {
		return "", nil
	}
	return s.CreateUserMessageFn(ctx, conversationID, content)
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
