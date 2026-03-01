package usecase

import (
	"context"
	"testing"

	chat "github.com/usetero/cli/internal/boundary/chat"
	"github.com/usetero/cli/internal/boundary/chat/chattest"
	corechat "github.com/usetero/cli/internal/core/chat"
	"github.com/usetero/cli/internal/domain"
)

func TestChatBoundaryGateway_StreamSnapshots_MapsRequestAndForwards(t *testing.T) {
	t.Parallel()

	var captured chat.Request
	wantMsg := &domain.Message{ID: "asst-1"}
	client := &chattest.MockClient{
		StreamSnapshotsFunc: func(_ context.Context, req chat.Request, onSnapshot func(corechat.StreamSnapshot)) (*corechat.StreamResult, error) {
			captured = req
			onSnapshot(corechat.StreamSnapshot{
				ConversationID: req.ConversationID,
				TurnID:         "turn-1",
				Seq:            1,
				Status:         corechat.StreamStatusCompleted,
				Done:           true,
				Message:        wantMsg,
			})
			return &corechat.StreamResult{Message: wantMsg}, nil
		},
	}

	gateway := NewChatBoundaryGateway(client)
	var gotSnapshots []corechat.StreamSnapshot
	result, err := gateway.StreamSnapshots(context.Background(), StreamRequest{
		ConversationID: "conv-1",
		Messages:       []domain.Message{{ID: "user-1", Role: domain.RoleUser}},
	}, func(s corechat.StreamSnapshot) {
		gotSnapshots = append(gotSnapshots, s)
	})
	if err != nil {
		t.Fatalf("StreamSnapshots() error = %v", err)
	}
	if captured.ConversationID != "conv-1" {
		t.Fatalf("conversation_id = %q, want conv-1", captured.ConversationID)
	}
	if len(captured.Messages) != 1 || captured.Messages[0].ID != "user-1" {
		t.Fatalf("messages mapping mismatch: %+v", captured.Messages)
	}
	if len(gotSnapshots) != 1 || gotSnapshots[0].Message != wantMsg {
		t.Fatalf("snapshot forwarding mismatch: %+v", gotSnapshots)
	}
	if result == nil || result.Message != wantMsg {
		t.Fatalf("result forwarding mismatch: %+v", result)
	}
}

func TestChatBoundaryGateway_StreamSnapshots_NilGatewayOrClient(t *testing.T) {
	t.Parallel()

	var nilGateway *ChatBoundaryGateway
	result, err := nilGateway.StreamSnapshots(context.Background(), StreamRequest{ConversationID: "conv-1"}, nil)
	if err != nil || result != nil {
		t.Fatalf("nil gateway = (%v, %v), want (nil, nil)", result, err)
	}

	gateway := NewChatBoundaryGateway(nil)
	result, err = gateway.StreamSnapshots(context.Background(), StreamRequest{ConversationID: "conv-1"}, nil)
	if err != nil || result != nil {
		t.Fatalf("nil client = (%v, %v), want (nil, nil)", result, err)
	}
}
