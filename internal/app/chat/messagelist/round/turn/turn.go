package turn

import (
	"context"
	"errors"
	"time"

	msgs "github.com/usetero/cli/internal/app/chat/events"
	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/user"
	"github.com/usetero/cli/internal/app/chat/usecase"
	chattools "github.com/usetero/cli/internal/app/chattools"
	corechat "github.com/usetero/cli/internal/core/chat"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
)

// State represents the current state of a turn.
type State int

const (
	StateIdle State = iota
	StateStreaming
	StateAwaitingToolResults
	StateComplete
)

const dbOpTimeout = 2 * time.Second

// Model represents a single user→assistant exchange.
// It is a fixed-height component - height is determined by content.
type Model struct {
	theme styles.Theme
	scope log.Scope

	conversationID domain.ConversationID
	accountID      domain.AccountID

	userMessage      *user.Model
	assistantMessage *assistant.Model

	state  State
	width  int
	stream *streamState

	// Tool result lifecycle (pending IDs/results/persisted/fired gate).
	toolTracker toolResultTracker

	// Protocol guard telemetry (incremented on dropped/malformed lifecycle events).
	protocolViolationCount int

	streamRunner       usecase.StreamRunner
	streamErrorMapper  usecase.StreamErrorMapper
	assistantPersister usecase.AssistantPersister
	toolRegistry       *chattools.Registry
}

// streamState holds the channel for receiving stream updates.
type streamState struct {
	updates chan streamUpdate
	cancel  context.CancelCauseFunc
	done    bool
}

// streamUpdate is sent through the channel as the stream progresses.
type streamUpdate struct {
	message *domain.Message
	status  corechat.StreamStatus
	abort   string
	result  *corechat.StreamResult // final result, only set on done
	err     error
	done    bool
}

var errUserCancelled = errors.New("user_cancelled")

// streamUpdateMsg is the internal message for stream handling.
type streamUpdateMsg struct {
	turnID domain.MessageID
	update streamUpdate
}

// assistantPersisted is fired after the assistant message is written to the DB.
// It gates fireToolResults to prevent racing with persistence.
type assistantPersisted struct {
	turnID    domain.MessageID
	messageID domain.MessageID
}

// New creates a new turn from a user submission.
func New(
	theme styles.Theme,
	conversationID domain.ConversationID,
	accountID domain.AccountID,
	userMessageID domain.MessageID,
	input msgs.UserSubmittedInput,
	width int,
	streamRunner usecase.StreamRunner,
	streamErrorMapper usecase.StreamErrorMapper,
	assistantPersister usecase.AssistantPersister,
	toolRegistry *chattools.Registry,
	scope log.Scope,
) *Model {
	scope = scope.Child("turn")
	return &Model{
		theme:              theme,
		scope:              scope,
		conversationID:     conversationID,
		accountID:          accountID,
		userMessage:        user.New(theme.WithBg(theme.BgElevated), userMessageID, input, width-block.BorderWidth),
		assistantMessage:   assistant.New(theme, userMessageID, "", width, toolRegistry, scope),
		state:              StateIdle,
		width:              width,
		streamRunner:       streamRunner,
		streamErrorMapper:  streamErrorMapper,
		assistantPersister: assistantPersister,
		toolRegistry:       toolRegistry,
	}
}
