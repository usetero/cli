package usecase

import (
	"context"
	"errors"
	"testing"

	corechat "github.com/usetero/cli/internal/core/chat"
	"github.com/usetero/cli/internal/domain"
)

type fakeGateway struct {
	streamSnapshots func(ctx context.Context, req StreamRequest, onSnapshot func(corechat.StreamSnapshot)) (*corechat.StreamResult, error)
}

func (f *fakeGateway) StreamSnapshots(ctx context.Context, req StreamRequest, onSnapshot func(corechat.StreamSnapshot)) (*corechat.StreamResult, error) {
	if f.streamSnapshots != nil {
		return f.streamSnapshots(ctx, req, onSnapshot)
	}
	return &corechat.StreamResult{}, nil
}

func TestChatStreamRunner_Start_ForwardsSnapshotsAndFinalResult(t *testing.T) {
	t.Parallel()

	reqs := make([]StreamRequest, 0, 1)
	streamMsg := &domain.Message{ID: "asst-1", Role: domain.RoleAssistant}
	finalResult := &corechat.StreamResult{
		Message: streamMsg,
		Metadata: &corechat.StreamMetadata{
			Title:         "Hello",
			ContextWindow: 200000,
			InputTokens:   12,
			OutputTokens:  4,
		},
	}

	gateway := &fakeGateway{
		streamSnapshots: func(_ context.Context, req StreamRequest, onSnapshot func(corechat.StreamSnapshot)) (*corechat.StreamResult, error) {
			reqs = append(reqs, req)
			onSnapshot(corechat.StreamSnapshot{
				ConversationID: req.ConversationID.String(),
				TurnID:         "turn-1",
				Seq:            1,
				Status:         corechat.StreamStatusStreaming,
				Message:        streamMsg,
			})
			onSnapshot(corechat.StreamSnapshot{
				ConversationID: req.ConversationID.String(),
				TurnID:         "turn-1",
				Seq:            2,
				Status:         corechat.StreamStatusCompleted,
				Done:           true,
				Message:        streamMsg,
				Metadata:       finalResult.Metadata,
			})
			return finalResult, nil
		},
	}

	runner := NewChatStreamRunner(gateway)
	updates := runner.Start(context.Background(), StreamRequest{
		ConversationID: "conv-1",
		Messages:       []domain.Message{{ID: "user-1", Role: domain.RoleUser}},
	})

	var got []StreamUpdate
	for u := range updates {
		got = append(got, u)
	}
	if len(got) != 2 {
		t.Fatalf("updates = %d, want 2", len(got))
	}
	if got[0].Done {
		t.Fatal("first update should be non-terminal")
	}
	if got[0].Status != corechat.StreamStatusStreaming {
		t.Fatalf("first status = %q, want %q", got[0].Status, corechat.StreamStatusStreaming)
	}
	if !got[1].Done {
		t.Fatal("last update should be terminal")
	}
	if got[1].Status != corechat.StreamStatusCompleted {
		t.Fatalf("last status = %q, want %q", got[1].Status, corechat.StreamStatusCompleted)
	}
	if got[1].Result == nil || got[1].Result.Metadata == nil || got[1].Result.Metadata.Title != "Hello" {
		t.Fatalf("terminal result metadata missing/invalid: %+v", got[1].Result)
	}
	if len(reqs) != 1 || reqs[0].ConversationID != "conv-1" {
		t.Fatalf("request mapping mismatch: %+v", reqs)
	}
}

func TestChatStreamRunner_Start_EmitsTerminalErrorUpdate(t *testing.T) {
	t.Parallel()

	gateway := &fakeGateway{
		streamSnapshots: func(_ context.Context, _ StreamRequest, _ func(corechat.StreamSnapshot)) (*corechat.StreamResult, error) {
			return nil, errors.New("network timeout")
		},
	}

	runner := NewChatStreamRunner(gateway)
	updates := runner.Start(context.Background(), StreamRequest{ConversationID: "conv-1"})
	var got []StreamUpdate
	for u := range updates {
		got = append(got, u)
	}
	if len(got) != 1 {
		t.Fatalf("updates = %d, want 1", len(got))
	}
	if !got[0].Done {
		t.Fatal("error update must be terminal")
	}
	if got[0].Err == nil || got[0].Err.Error() != "network timeout" {
		t.Fatalf("err = %v, want network timeout", got[0].Err)
	}
}

func TestChatStreamRunner_Start_NilClientClosesImmediately(t *testing.T) {
	t.Parallel()

	var runner *ChatStreamRunner
	updates := runner.Start(context.Background(), StreamRequest{ConversationID: "conv-1"})
	if _, ok := <-updates; ok {
		t.Fatal("expected closed updates channel for nil runner")
	}
}
