package chat

import (
	"fmt"

	"github.com/usetero/cli/internal/domain"
)

type StreamStatus string

const (
	StreamStatusStreaming StreamStatus = "streaming"
	StreamStatusCompleted StreamStatus = "completed"
	StreamStatusToolUse   StreamStatus = "tool_use"
	StreamStatusAborted   StreamStatus = "aborted"
)

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

type reducer struct {
	acc            *accumulator
	conversationID string
	turnID         string
	lastSeq        int
	terminal       bool
	started        bool
	stopped        bool
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

	if e.Done {
		if !r.stopped {
			return nil, fmt.Errorf("protocol error: received [DONE] before message_stop")
		}
		if err := r.acc.handle(e); err != nil {
			return nil, err
		}
		r.terminal = true

		status := StreamStatusCompleted
		if r.acc.stopReason == "tool_use" {
			status = StreamStatusToolUse
		}

		return &StreamSnapshot{
			ConversationID: r.conversationID,
			TurnID:         r.turnID,
			Seq:            r.lastSeq,
			Status:         status,
			Done:           true,
			Message:        r.acc.message(),
			Metadata:       r.metadata(),
		}, nil
	}

	if !r.started {
		if e.Type != EventTypeMessageStart {
			return nil, fmt.Errorf("protocol error: first event must be message_start, got %q", e.Type)
		}
		r.started = true
	} else {
		if e.Type == EventTypeMessageStart {
			return nil, fmt.Errorf("protocol error: duplicate message_start")
		}
		if r.stopped && e.Type != EventTypeMetadataUpdate {
			return nil, fmt.Errorf("protocol error: event %q after message_stop", e.Type)
		}
	}

	if err := r.acc.handle(e); err != nil {
		return nil, err
	}

	if e.Type == EventTypeMessageStop {
		r.stopped = true
	}

	return &StreamSnapshot{
		ConversationID: r.conversationID,
		TurnID:         r.turnID,
		Seq:            r.lastSeq,
		Status:         StreamStatusStreaming,
		Done:           false,
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
