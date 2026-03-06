package chat

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	messagesdb "github.com/usetero/cli/internal/domains/chat/db/messagesgen"
)

// LocalMessageService uses SQLite/sqlc for message CRUD.
type LocalMessageService struct {
	q *messagesdb.Queries
}

func NewLocalMessageService(db *sql.DB) *LocalMessageService {
	if db == nil {
		panic("chat local message service requires db")
	}
	return &LocalMessageService{q: messagesdb.New(db)}
}

// CreateUserMessage inserts a user-authored message.
func (s *LocalMessageService) CreateUserMessage(ctx context.Context, create UserMessageCreate) (MessageID, error) {
	validated, err := create.Validate()
	if err != nil {
		return "", err
	}

	id := MessageID(uuid.NewString())
	err = s.q.Create(ctx, messagesdb.CreateParams{
		ID:             toMessagesDBMessageID(id),
		ConversationID: toMessagesDBConversationID(validated.ConversationID),
		Role:           toMessagesDBRole(RoleUser),
		Content:        validated.Content,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

// Delete removes a message by id.
func (s *LocalMessageService) Delete(ctx context.Context, messageID MessageID) error {
	if messageID == "" {
		return fmt.Errorf("message id is required")
	}
	return s.q.Delete(ctx, toMessagesDBMessageID(messageID))
}

// ListByConversation returns messages for one conversation ordered by creation time.
func (s *LocalMessageService) ListByConversation(ctx context.Context, conversationID ConversationID) ([]Message, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("conversation id is required")
	}
	rows, err := s.q.ListByConversation(ctx, toMessagesDBConversationID(conversationID))
	if err != nil {
		return nil, err
	}

	var out []Message
	for _, row := range rows {
		out = append(out, fromMessagesDBMessage(row))
	}
	return out, nil
}
