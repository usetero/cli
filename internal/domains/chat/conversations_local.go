package chat

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	conversationsdb "github.com/usetero/cli/internal/domains/chat/db/conversationsgen"
)

// LocalConversationService uses SQLite/sqlc for conversation CRUD.
type LocalConversationService struct {
	q *conversationsdb.Queries
}

func NewLocalConversationService(db *sql.DB) *LocalConversationService {
	if db == nil {
		panic("chat local conversation service requires db")
	}
	return &LocalConversationService{q: conversationsdb.New(db)}
}

// Create inserts a conversation.
func (s *LocalConversationService) Create(ctx context.Context, create ConversationCreate) (ConversationID, error) {
	validated, err := create.Validate()
	if err != nil {
		return "", err
	}
	id := ConversationID(uuid.NewString())
	err = s.q.Create(ctx, conversationsdb.CreateParams{
		ID:        toConversationsDBConversationID(id),
		Title:     validated.Title,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

// Delete removes a conversation by id.
func (s *LocalConversationService) Delete(ctx context.Context, id ConversationID) error {
	if id == "" {
		return fmt.Errorf("conversation id is required")
	}
	return s.q.Delete(ctx, toConversationsDBConversationID(id))
}

// List returns conversations ordered by creation time descending.
func (s *LocalConversationService) List(ctx context.Context) ([]Conversation, error) {
	rows, err := s.q.List(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]Conversation, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromConversationsDBConversation(row))
	}
	return out, nil
}
