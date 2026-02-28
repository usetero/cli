package chat

import "github.com/usetero/cli/internal/domain"

type StreamStatus string

const (
	StreamStatusStreaming StreamStatus = "streaming"
	StreamStatusCompleted StreamStatus = "completed"
	StreamStatusToolUse   StreamStatus = "tool_use"
	StreamStatusAborted   StreamStatus = "aborted"
)

// StreamMetadata contains post-stream metadata from the Chat API.
type StreamMetadata struct {
	Title         string
	ContextWindow int
	InputTokens   int
	OutputTokens  int
}

// StreamSnapshot is a deterministic progress snapshot for one stream turn.
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

// StreamMachine consumes SSE data payloads and emits typed stream snapshots.
type StreamMachine struct {
	red *reducer
}

func NewStreamMachine(conversationID string) *StreamMachine {
	return &StreamMachine{red: newReducer(conversationID)}
}

// ConsumeData parses and applies one non-[DONE] SSE data payload.
func (m *StreamMachine) ConsumeData(data []byte) (*StreamSnapshot, error) {
	e, err := decodeEventData(data)
	if err != nil {
		return nil, err
	}
	return m.red.apply(e)
}

// ConsumeDone applies terminal [DONE].
func (m *StreamMachine) ConsumeDone() (*StreamSnapshot, error) {
	return m.red.apply(event{Done: true})
}

// Abort emits a terminal aborted snapshot.
func (m *StreamMachine) Abort(reason string) *StreamSnapshot {
	return m.red.abortSnapshot(reason)
}

func (m *StreamMachine) Message() *domain.Message {
	return m.red.acc.message()
}

func (m *StreamMachine) Metadata() *StreamMetadata {
	return m.red.metadata()
}

