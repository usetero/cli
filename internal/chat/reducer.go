package chat

import (
	"fmt"

	"github.com/usetero/cli/internal/domain"
)

// StreamStatus is the normalized lifecycle status for a chat turn stream.
type StreamStatus string

const (
	StreamStatusStreaming StreamStatus = "streaming"
	StreamStatusCompleted StreamStatus = "completed"
	StreamStatusToolUse   StreamStatus = "tool_use"
	StreamStatusAborted   StreamStatus = "aborted"
)

// StreamSnapshot is the reducer output after applying one stream event.
// It gives callers a typed, turn-scoped view of stream progress.
type StreamSnapshot struct {
	ConversationID string
	TurnID         string
	Seq            int
	Status         StreamStatus
	AbortReason    string
	Done           bool
	Message        *domain.Message
	Metadata       *StreamMetadata
}

// reducer validates ordering/scoping and builds snapshots from wire events.
type reducer struct {
	acc            *accumulator
	conversationID string
	turnID         string
	lastSeq        int
	terminal       bool
}

func newReducer(conversationID string) *reducer {
	return &reducer{
		acc:            newAccumulator(),
		conversationID: conversationID,
	}
}

func (r *reducer) apply(e event) (*StreamSnapshot, error) {
	if r.terminal {
		return nil, fmt.Errorf("received event after terminal state")
	}

	if e.ConversationID != "" && r.conversationID != "" && e.ConversationID != r.conversationID {
		return nil, fmt.Errorf("conversation_id mismatch: got %q want %q", e.ConversationID, r.conversationID)
	}
	if e.ConversationID != "" {
		r.conversationID = e.ConversationID
	}

	if e.TurnID != "" {
		if r.turnID == "" {
			r.turnID = e.TurnID
		} else if e.TurnID != r.turnID {
			return nil, fmt.Errorf("turn_id mismatch: got %q want %q", e.TurnID, r.turnID)
		}
	}

	if e.Seq > 0 {
		if r.lastSeq > 0 && e.Seq <= r.lastSeq {
			return nil, fmt.Errorf("non-monotonic seq: got %d after %d", e.Seq, r.lastSeq)
		}
		r.lastSeq = e.Seq
	}

	r.acc.handle(e)

	status := StreamStatusStreaming
	if e.Done {
		r.terminal = true
		if r.acc.stopReason == "tool_use" {
			status = StreamStatusToolUse
		} else {
			status = StreamStatusCompleted
		}
	}

	return &StreamSnapshot{
		ConversationID: r.conversationID,
		TurnID:         r.turnID,
		Seq:            r.lastSeq,
		Status:         status,
		Done:           e.Done,
		Message:        r.acc.message(),
		Metadata:       r.metadata(),
	}, nil
}

func (r *reducer) abortSnapshot(reason string) *StreamSnapshot {
	if r.terminal {
		return nil
	}
	r.terminal = true
	return &StreamSnapshot{
		ConversationID: r.conversationID,
		TurnID:         r.turnID,
		Seq:            r.lastSeq,
		Status:         StreamStatusAborted,
		AbortReason:    reason,
		Done:           true,
		Message:        r.acc.message(),
		Metadata:       r.metadata(),
	}
}

func (r *reducer) metadata() *StreamMetadata {
	if r.acc.title == "" && r.acc.contextWindow == 0 && r.acc.inputTokens == 0 && r.acc.outputTokens == 0 {
		return nil
	}
	return &StreamMetadata{
		Title:         r.acc.title,
		ContextWindow: r.acc.contextWindow,
		InputTokens:   r.acc.inputTokens,
		OutputTokens:  r.acc.outputTokens,
	}
}
