package usecase

import (
	"context"

	chatclient "github.com/usetero/cli/internal/api/chatclient"
	corechat "github.com/usetero/cli/internal/core/chat"
	"github.com/usetero/cli/internal/domain"
)

// StreamUpdate is a normalized stream event consumed by UI models.
type StreamUpdate struct {
	Message     *domain.Message
	Status      corechat.StreamStatus
	AbortReason string
	Result      *corechat.StreamResult
	Err         error
	Done        bool
}

// StreamRunner executes one chat stream request and emits ordered updates.
type StreamRunner interface {
	Start(ctx context.Context, req chatclient.Request) <-chan StreamUpdate
}

// ChatStreamRunner bridges chatclient snapshots into use-case updates.
type ChatStreamRunner struct {
	client chatclient.Client
}

func NewChatStreamRunner(client chatclient.Client) *ChatStreamRunner {
	return &ChatStreamRunner{client: client}
}

func (r *ChatStreamRunner) Start(ctx context.Context, req chatclient.Request) <-chan StreamUpdate {
	updates := make(chan StreamUpdate, 10)
	if r == nil || r.client == nil {
		close(updates)
		return updates
	}

	go func() {
		defer close(updates)

		var lastSnapshot *corechat.StreamSnapshot
		result, err := r.client.StreamSnapshots(ctx, req, func(s corechat.StreamSnapshot) {
			ss := s
			lastSnapshot = &ss
			if !s.Done {
				updates <- StreamUpdate{Message: s.Message, Status: s.Status}
			}
		})
		if err != nil {
			updates <- StreamUpdate{Err: err, Done: true}
			return
		}

		var finalMsg *domain.Message
		var status corechat.StreamStatus
		var abort string
		if lastSnapshot != nil {
			finalMsg = lastSnapshot.Message
			status = lastSnapshot.Status
			abort = lastSnapshot.AbortReason
		}
		updates <- StreamUpdate{
			Message:     finalMsg,
			Status:      status,
			AbortReason: abort,
			Result:      result,
			Done:        true,
		}
	}()

	return updates
}
