package round

import (
	"time"

	tea "charm.land/bubbletea/v2"
	chatclient "github.com/usetero/cli/internal/api/chatclient"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn"
	"github.com/usetero/cli/internal/app/chat/msgs"
	chattools "github.com/usetero/cli/internal/app/chattools"
	corechat "github.com/usetero/cli/internal/core/chat"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/thinking"
)

// State represents the current state of a round.
type State int

const (
	StateActive           State = iota
	StateAwaitingNextTurn       // async DB work in progress before next turn
	StateComplete
	StateCancelled
	StateFailed
)

const dbOpTimeout = 2 * time.Second

// IsActive returns true if the round is in-flight (active or awaiting next turn).
func (m *Model) IsActive() bool {
	return m.state == StateActive || m.state == StateAwaitingNextTurn
}

// Model represents a complete user→assistant exchange, potentially with multiple turns
// if tools are involved. A round starts with explicit user input and ends when the
// assistant stops (no more tool calls).
// It is a fixed-height component - height is determined by content.
type Model struct {
	theme styles.Theme
	scope log.Scope

	id             domain.MessageID // first user message ID, identifies the round
	conversationID domain.ConversationID
	accountID      domain.AccountID

	turns    []*turn.Model
	session  *corechat.Session // authoritative in-memory history for active tool loop
	thinking *thinking.Model
	state    State
	lastErr  error
	width    int

	startTime time.Time
	endTime   time.Time

	db           sqlite.DB
	chatClient   chatclient.Client
	toolRegistry *chattools.Registry
}

// New creates a new round from explicit user input.
func New(
	theme styles.Theme,
	conversationID domain.ConversationID,
	accountID domain.AccountID,
	userMessageID domain.MessageID,
	input msgs.UserSubmittedInput,
	width int,
	db sqlite.DB,
	chatClient chatclient.Client,
	toolRegistry *chattools.Registry,
	scope log.Scope,
) *Model {
	scope = scope.Child("round")

	// Create first turn with user's explicit input
	firstTurn := turn.New(
		theme,
		conversationID,
		accountID,
		userMessageID,
		input,
		width,
		db,
		chatClient,
		toolRegistry,
		scope,
	)

	return &Model{
		theme:          theme,
		scope:          scope,
		id:             userMessageID,
		conversationID: conversationID,
		accountID:      accountID,
		turns:          []*turn.Model{firstTurn},
		thinking:       thinking.New(theme, thinking.Settings{Label: "Thinking"}),
		state:          StateActive,
		width:          width,
		startTime:      time.Now(),
		db:             db,
		chatClient:     chatClient,
		toolRegistry:   toolRegistry,
	}
}

// Init starts the thinking animation.
func (m *Model) Init() tea.Cmd {
	return m.thinking.Init()
}

// StartStream begins streaming for the first turn.
func (m *Model) StartStream(messages []domain.Message, context []domain.ContextEntity) tea.Cmd {
	if len(m.turns) == 0 {
		return nil
	}
	m.session = corechat.NewSession(m.conversationID, messages)
	return m.turns[0].StartStream(messages, context)
}
