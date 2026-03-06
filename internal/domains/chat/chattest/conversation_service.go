package chattest

import (
	"context"

	domainchat "github.com/usetero/cli/internal/domains/chat"
)

type ConversationService struct {
	CreateFn func(ctx context.Context, create domainchat.ConversationCreate) (domainchat.ConversationID, error)
	DeleteFn func(ctx context.Context, id domainchat.ConversationID) error
	ListFn   func(ctx context.Context) ([]domainchat.Conversation, error)
}

var _ domainchat.ConversationService = (*ConversationService)(nil)

func (s *ConversationService) Create(ctx context.Context, create domainchat.ConversationCreate) (domainchat.ConversationID, error) {
	if s.CreateFn == nil {
		return "", nil
	}
	return s.CreateFn(ctx, create)
}

func (s *ConversationService) Delete(ctx context.Context, id domainchat.ConversationID) error {
	if s.DeleteFn == nil {
		return nil
	}
	return s.DeleteFn(ctx, id)
}

func (s *ConversationService) List(ctx context.Context) ([]domainchat.Conversation, error) {
	if s.ListFn == nil {
		return nil, nil
	}
	return s.ListFn(ctx)
}
