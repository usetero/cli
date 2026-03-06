package chat

import (
	"context"

	domainchat "github.com/usetero/cli/internal/domains/chat"
)

func (r *Runtime) ensureConversation(ctx context.Context) (domainchat.ConversationID, error) {
	r.mu.RLock()
	if r.state.ConversationID != "" {
		id := r.state.ConversationID
		r.mu.RUnlock()
		return id, nil
	}
	r.mu.RUnlock()

	id, err := r.conversations.Create(ctx, domainchat.ConversationCreate{})
	if err != nil {
		return "", err
	}

	r.mu.Lock()
	r.state.ConversationID = id
	snapshot := cloneState(r.state)
	r.mu.Unlock()
	r.publish(snapshot)
	return id, nil
}

func (r *Runtime) persistUserMessage(ctx context.Context, conversationID domainchat.ConversationID, text string) (domainchat.MessageID, error) {
	id, err := r.messages.CreateUserMessage(ctx, domainchat.UserMessageCreate{
		ConversationID: conversationID,
		Content:        text,
	})
	if err != nil {
		return "", err
	}
	return id, nil
}
