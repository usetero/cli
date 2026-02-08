package chat

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/chat/inputbar"
	"github.com/usetero/cli/internal/app/chat/messagelist"
	"github.com/usetero/cli/internal/app/chat/msgs"
	appmsg "github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/auth"
	chatclient "github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/keymap"
)

// Chat-specific key bindings.
var (
	scrollUp = key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑↓", "scroll"),
	)
	focusInputBar = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "focus input"),
	)
	focusChat = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "focus chat"),
	)
)

// focus tracks which component has keyboard focus.
type focus int

const (
	focusEditor focus = iota
	focusMessages
)

// Model is the main chat model.
// It is a flexible component - it renders exactly the size given by SetSize.
type Model struct {
	scope log.Scope
	focus focus

	inputBar    *inputbar.Model
	messageList *messagelist.Model

	// Conversation is created lazily on first message
	conversationID domain.ConversationID

	user      *auth.User
	account   domain.Account
	workspace domain.Workspace
	theme     styles.Theme
	width     int
	height    int
	originX   int
	originY   int

	// Empty state
	policySummary *domain.PolicySummary

	// Dependencies
	db           sqlite.DB
	chatClient   chatclient.Client
	toolRegistry *tools.Registry
}

// emptyStatePollMsg triggers a policy summary fetch for the empty state.
type emptyStatePollMsg struct{}

// New creates a new chat model.
func New(
	user *auth.User,
	account domain.Account,
	workspace domain.Workspace,
	theme styles.Theme,
	db sqlite.DB,
	chatClient chatclient.Client,
	toolRegistry *tools.Registry,
	scope log.Scope,
) *Model {
	scope = scope.Child("chat")

	return &Model{
		scope:        scope,
		inputBar:     inputbar.New(user, theme, scope),
		messageList:  messagelist.New(theme, db, chatClient, toolRegistry, scope),
		user:         user,
		account:      account,
		workspace:    workspace,
		theme:        theme,
		db:           db,
		chatClient:   chatClient,
		toolRegistry: toolRegistry,
	}
}

// Init initializes the model.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.inputBar.Init(),
		m.pollEmptyState(),
	)
}

func (m *Model) pollEmptyState() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return emptyStatePollMsg{}
	})
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Handle messages this model cares about
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, keymap.Tab) {
			cmds = append(cmds, m.toggleFocus())
			return tea.Batch(cmds...)
		}
		if m.focus == focusMessages {
			// Enter or esc returns to editor
			if key.Matches(msg, keymap.Send) || key.Matches(msg, keymap.Exit) {
				cmds = append(cmds, m.setFocus(focusEditor))
				return tea.Batch(cmds...)
			}
			// Only forward to message list when it's focused
			cmds = append(cmds, m.messageList.Update(msg))
			return tea.Batch(cmds...)
		}

	case emptyStatePollMsg:
		if !m.hasMessages() && m.db != nil {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			summary, err := m.db.DatadogAccountStatuses().GetPolicySummary(ctx)
			if err == nil && summary.ReadyForUse {
				m.policySummary = &summary
			}
			return m.pollEmptyState()
		}
		return nil // stop polling once messages exist

	case msgs.UserSubmittedInput:
		cmds = append(cmds, m.handleUserInput(msg))

	case conversationCreated:
		m.scope.Info("conversation created", "id", msg.conversationID)
		m.conversationID = msg.conversationID
		cmds = append(cmds, m.persistUserMessage(msg.input))

	case userMessagePersisted:
		m.scope.Debug("received userMessagePersisted", "message_id", msg.messageID)
		cmds = append(cmds, m.handlePersistedMessage(msg))

	case tea.MouseClickMsg:
		// Click on the message list area focuses it
		if m.hasMessages() && msg.Y >= m.originY && msg.Y < m.originY+m.height-m.inputBar.Height() {
			if m.focus != focusMessages {
				cmds = append(cmds, m.setFocus(focusMessages))
			}
		} else if msg.Y >= m.originY+m.height-m.inputBar.Height() {
			if m.focus != focusEditor {
				cmds = append(cmds, m.setFocus(focusEditor))
			}
		}
	}

	// Forward to children
	cmds = append(cmds, m.inputBar.Update(msg))
	cmds = append(cmds, m.messageList.Update(msg))

	return tea.Batch(cmds...)
}

// toggleFocus switches focus between editor and messages.
func (m *Model) toggleFocus() tea.Cmd {
	switch m.focus {
	case focusEditor:
		if !m.hasMessages() {
			return nil // nothing to focus
		}
		return m.setFocus(focusMessages)
	default:
		return m.setFocus(focusEditor)
	}
}

// setFocus sets focus to the given target.
func (m *Model) setFocus(f focus) tea.Cmd {
	m.focus = f
	switch f {
	case focusEditor:
		m.messageList.SetFocused(false)
		return m.inputBar.Focus()
	case focusMessages:
		m.inputBar.Blur()
		m.messageList.SetFocused(true)
		return nil
	}
	return nil
}

// SetSize updates the dimensions. This is a flexible component.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.updateLayout()
}

// SetOrigin sets the terminal-absolute position of this component's top-left corner.
func (m *Model) SetOrigin(x, y int) {
	m.originX = x
	m.originY = y
	m.updateLayout()
}

// updateLayout calculates sizes for children based on current dimensions.
func (m *Model) updateLayout() {
	// Input bar is fixed height
	m.inputBar.SetWidth(m.width)
	inputBarHeight := m.inputBar.Height()

	// MessageList is flexible - gets remaining space
	messageListHeight := m.height - inputBarHeight
	if messageListHeight < 0 {
		messageListHeight = 0
	}
	m.messageList.SetSize(m.width, messageListHeight)
	m.messageList.SetOrigin(m.originX, m.originY)
}

// ShortHelp returns the key bindings for the short help view.
func (m *Model) ShortHelp() []key.Binding {
	if m.focus == focusMessages {
		return []key.Binding{scrollUp, focusInputBar}
	}
	if m.hasMessages() {
		return append(m.inputBar.ShortHelp(), focusChat)
	}
	return m.inputBar.ShortHelp()
}

// ConversationID returns the current conversation ID.
func (m *Model) ConversationID() domain.ConversationID {
	return m.conversationID
}

// CancelActiveRound cancels the active round if one exists.
// Returns true if a round was cancelled.
func (m *Model) CancelActiveRound() bool {
	if !m.messageList.HasActiveRound() {
		return false
	}
	m.messageList.CancelActiveRound()
	return true
}

// hasMessages returns true if there are messages to display.
func (m *Model) hasMessages() bool {
	return m.messageList.Len() > 0
}

// handleUserInput creates conversation if needed, then persists the user message.
func (m *Model) handleUserInput(input msgs.UserSubmittedInput) tea.Cmd {
	if len(input.Text) > 0 {
		m.scope.Info("user submitted text", "text_length", len(input.Text))
	} else {
		m.scope.Info("user submitted tool results", "count", len(input.ToolResults))
	}

	// Cancel any in-flight round before starting a new one.
	m.messageList.CancelActiveRound()

	// If no conversation yet, create one first (only for text input)
	if m.conversationID == "" {
		return m.createConversation(input)
	}

	return m.persistUserMessage(input)
}

// createConversation creates a new conversation.
func (m *Model) createConversation(input msgs.UserSubmittedInput) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		convID, err := m.db.Conversations().Create(
			ctx,
			m.account.ID.String(),
			m.workspace.ID.String(),
		)
		if err != nil {
			m.scope.Error("failed to create conversation", "error", err)
			return appmsg.Error{Message: "Failed to create conversation", Err: err}
		}

		return conversationCreated{
			conversationID: domain.ConversationID(convID),
			input:          input,
		}
	}
}

// conversationCreated is fired after conversation is created.
type conversationCreated struct {
	conversationID domain.ConversationID
	input          msgs.UserSubmittedInput
}

// persistUserMessage saves the user message and loads history for the API call.
func (m *Model) persistUserMessage(input msgs.UserSubmittedInput) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var msgID domain.MessageID
		var err error

		if len(input.ToolResults) > 0 {
			// Convert typed results to domain format at the boundary
			domainResults := make([]domain.ToolResult, len(input.ToolResults))
			for i, r := range input.ToolResults {
				domainResults[i] = domain.ToolResult{
					ToolUseID: r.ToolUseID,
					IsError:   r.IsError(),
					Content:   r.ToMap(),
				}
				if r.Error != nil {
					domainResults[i].Error = r.Error.Message
				}
			}
			msgID, err = m.db.Messages().CreateToolResultMessage(ctx, m.account.ID, m.conversationID, domainResults)
		} else {
			msgID, err = m.db.Messages().CreateUserMessage(ctx, m.account.ID, m.conversationID, input.Text)
		}
		if err != nil {
			m.scope.Error("failed to create user message", "error", err)
			return appmsg.Error{Message: "Failed to save message", Err: err}
		}

		messages, err := m.db.Messages().List(ctx, m.conversationID)
		if err != nil {
			m.scope.Error("failed to load messages", "error", err)
			return appmsg.Error{Message: "Failed to load messages", Err: err}
		}

		return userMessagePersisted{
			conversationID: m.conversationID,
			messageID:      msgID,
			input:          input,
			messages:       messages,
		}
	}
}

// userMessagePersisted is fired after user message is saved to database.
type userMessagePersisted struct {
	conversationID domain.ConversationID
	messageID      domain.MessageID
	input          msgs.UserSubmittedInput
	messages       []domain.Message
}

// handlePersistedMessage starts the turn after the user message is persisted.
func (m *Model) handlePersistedMessage(msg userMessagePersisted) tea.Cmd {
	m.scope.Info("turn started", "conversation_id", msg.conversationID, "user_message_id", msg.messageID)

	return m.messageList.StartTurn(
		msg.conversationID,
		m.account.ID,
		msg.messageID,
		msg.input,
		msg.messages,
		nil,
	)
}

// View renders the chat. This is a flexible component - renders exactly to SetSize dimensions.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	// Empty state: context-aware greeting + suggestions + input bar
	if !m.hasMessages() {
		emptyHeight := m.height - m.inputBar.Height()
		emptyView := lipgloss.NewStyle().
			Width(m.width).
			Height(emptyHeight).
			Align(lipgloss.Center, lipgloss.Center).
			Render(m.emptyStateContent())

		return lipgloss.JoinVertical(lipgloss.Left, emptyView, m.inputBar.View())
	}

	// Normal state: message list + input bar
	return lipgloss.JoinVertical(lipgloss.Left, m.messageList.View(), m.inputBar.View())
}

// emptyStateContent renders the context-aware empty state.
func (m *Model) emptyStateContent() string {
	colors := m.theme
	text := lipgloss.NewStyle().Foreground(colors.Text)
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted)

	name := ""
	if m.user != nil && m.user.FirstName != "" {
		name = m.user.FirstName
	}

	var headline string
	var suggestions []string

	s := m.policySummary
	if s == nil || !s.ReadyForUse {
		// No data yet — generic greeting
		if name != "" {
			headline = fmt.Sprintf("Hey %s — loading your environment...", name)
		} else {
			headline = "Loading your environment..."
		}
		suggestions = []string{
			"What services are generating the most logs?",
			"Show me a summary of my environment",
			"What does Tero do?",
		}
	} else if s.PendingPolicyCount > 0 {
		// Pending work
		wp := wastePercent(*s)
		if name != "" {
			headline = fmt.Sprintf("Hey %s — %d%% waste across your services. Let's dig in:", name, wp)
		} else {
			headline = fmt.Sprintf("%d%% waste across your services. Let's dig in:", wp)
		}
		suggestions = []string{
			"Walk me through the pending policies",
			"Which policies should I approve first?",
			"Show me what's driving the most waste",
		}
	} else if s.ApprovedPolicyCount > 0 {
		// Mid-journey — approved some, none pending
		if name != "" {
			headline = fmt.Sprintf("Nice work, %s — %d policies approved. Let's keep going:", name, s.ApprovedPolicyCount)
		} else {
			headline = fmt.Sprintf("%d policies approved. Let's keep going:", s.ApprovedPolicyCount)
		}
		suggestions = []string{
			"How are the approved policies performing?",
			"Are there any new recommendations?",
			"Show me observed savings so far",
		}
	} else {
		// All clean — no pending, no approved
		if name != "" {
			headline = fmt.Sprintf("Looking good, %s — I'm watching for changes.", name)
		} else {
			headline = "Looking good — I'm watching for changes."
		}
		suggestions = []string{
			"What services are generating the most logs?",
			"Show me a summary of my environment",
			"Any optimization opportunities?",
		}
	}

	var lines []string
	lines = append(lines, text.Render(headline))
	lines = append(lines, "")
	for _, s := range suggestions {
		lines = append(lines, muted.Render("  → "+s))
	}

	return strings.Join(lines, "\n")
}

// wastePercent computes waste % preferring bytes.
func wastePercent(s domain.PolicySummary) int {
	if s.TotalBytesPerHour > 0 && s.EstimatedBytesPerHour > 0 {
		return int(math.Round(s.EstimatedBytesPerHour / s.TotalBytesPerHour * 100))
	}
	if s.EstimatedCostPerHour != nil && s.TotalCostPerHour != nil && *s.TotalCostPerHour > 0 {
		return int(math.Round(*s.EstimatedCostPerHour / *s.TotalCostPerHour * 100))
	}
	if s.TotalVolumePerHour > 0 && s.EstimatedVolumePerHour > 0 {
		return int(math.Round(s.EstimatedVolumePerHour / s.TotalVolumePerHour * 100))
	}
	return 0
}
