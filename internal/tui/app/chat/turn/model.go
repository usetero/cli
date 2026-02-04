package turn

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/thinking"
)

// StreamDoneMsg is sent when streaming completes.
// Contains the full assistant message.
type StreamDoneMsg struct {
	Message    domain.Message
	StopReason string
	Err        error
}

// Model handles API communication and displays a thinking indicator while active.
type Model struct {
	client   chat.Client
	logger   log.Logger
	theme    *styles.Theme
	thinking thinking.Model
	active   bool
}

// New creates a new Turn model.
func New(theme *styles.Theme, client chat.Client, logger log.Logger) Model {
	return Model{
		theme:  theme,
		client: client,
		logger: logger,
	}
}

// Update handles messages.
// Rule: return early ONLY if this model is the sole consumer of the message.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	// Handle messages we care about
	switch msg.(type) {
	case StreamDoneMsg:
		m.active = false // sole consumer (of deactivation)
	}

	// Forward to children
	var cmd tea.Cmd
	m.thinking, cmd = m.thinking.Update(msg)
	return m, cmd
}

// View renders the thinking indicator while active.
func (m Model) View() string {
	if !m.active {
		return ""
	}
	return m.thinking.View()
}

// IsActive returns true if Turn is currently streaming.
func (m Model) IsActive() bool {
	return m.active
}

// Send starts streaming a request to the Chat API.
// Returns a Cmd that blocks until streaming completes, then emits StreamDoneMsg.
func (m *Model) Send(ctx context.Context, conversationID string, messages []domain.Message, tools []chat.Tool) tea.Cmd {
	// Activate and initialize thinking indicator
	m.active = true
	m.thinking = thinking.New(m.theme, "Thinking")

	client := m.client
	logger := m.logger

	return tea.Batch(
		m.thinking.Init(),
		func() tea.Msg {
			req := chat.Request{
				ConversationID: conversationID,
				Messages:       messages,
				Tools:          tools,
			}

			logger.Info("streaming")
			logger.Debug("stream request",
				"conversation_id", conversationID,
				"messages", len(messages),
				"tools", len(tools),
			)

			var lastMessage *domain.Message

			err := client.Stream(ctx, req, func(msg *domain.Message) {
				lastMessage = msg
			})

			if err != nil {
				logger.Error("stream failed", "error", err)
				return StreamDoneMsg{Err: err}
			}

			if lastMessage == nil {
				logger.Error("stream completed with no message")
				return StreamDoneMsg{Err: err}
			}

			logger.Debug("stream complete",
				"stop_reason", lastMessage.StopReason,
				"blocks", len(lastMessage.Content),
			)

			return StreamDoneMsg{
				Message: domain.Message{
					ID:             domain.NewMessageID(),
					ConversationID: domain.ConversationID(conversationID),
					Role:           domain.RoleAssistant,
					Content:        lastMessage.Content,
					Model:          lastMessage.Model,
					StopReason:     lastMessage.StopReason,
				},
				StopReason: lastMessage.StopReason,
			}
		},
	)
}
