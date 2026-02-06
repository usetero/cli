package round

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn"
	"github.com/usetero/cli/internal/app/chat/msgs"
	chatclient "github.com/usetero/cli/internal/chat"
	chattools "github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/domain"
	domaintools "github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

// Layout constants for gaps (parent owns spacing between children).
const (
	gapBetweenTurns  = 1 // blank lines between turns in a multi-turn round
	gapBeforeDivider = 1 // blank lines before the completion divider
)

// State represents the current state of a round.
type State int

const (
	StateActive State = iota
	StateComplete
)

// Model represents a complete user→assistant exchange, potentially with multiple turns
// if tools are involved. A round starts with explicit user input and ends when the
// assistant stops (no more tool calls).
// It is a fixed-height component - height is determined by content.
type Model struct {
	theme *styles.Theme
	scope log.Scope

	id             domain.MessageID // first user message ID, identifies the round
	conversationID domain.ConversationID
	accountID      domain.AccountID

	turns     []*turn.Model
	state     State
	width     int
	startTime time.Time
	endTime   time.Time

	db           sqlite.DB
	chatClient   chatclient.Client
	toolRegistry *chattools.Registry
}

// New creates a new round from explicit user input.
func New(
	theme *styles.Theme,
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
		state:          StateActive,
		width:          width,
		startTime:      time.Now(),
		db:             db,
		chatClient:     chatClient,
		toolRegistry:   toolRegistry,
	}
}

// StartStream begins streaming for the first turn.
func (m *Model) StartStream(messages []domain.Message, context []domain.ContextEntity) tea.Cmd {
	if len(m.turns) == 0 {
		return nil
	}
	return m.turns[0].StartStream(messages, context)
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case msgs.StreamCompleted:
		// Check if this is for our current turn
		if m.isOurTurn(msg.TurnID) {
			if msg.StopReason != "tool_use" {
				// Round complete - no more tool calls
				m.state = StateComplete
				m.endTime = time.Now()
				m.scope.Info("round complete", "stop_reason", msg.StopReason)
			}
			// If tool_use, turn will collect results and fire ToolResultsReady
		}

	case msgs.ToolResultsReady:
		// Tool results collected by one of our turns - start next turn
		if m.isOurTurn(msg.TurnID) && m.state == StateActive {
			cmds = append(cmds, m.startNextTurn(msg.Results))
		}

	case nextTurnReady:
		// Internal message after persistence - create and start next turn
		if msg.roundID == m.id {
			cmds = append(cmds, m.handleNextTurnReady(msg))
		}
	}

	// Forward to all turns
	for _, t := range m.turns {
		cmds = append(cmds, t.Update(msg))
	}

	return tea.Batch(cmds...)
}

// isOurTurn checks if the given turn ID belongs to this round.
func (m *Model) isOurTurn(turnID domain.MessageID) bool {
	for _, t := range m.turns {
		if t.UserMessageID() == turnID {
			return true
		}
	}
	return false
}

// startNextTurn persists tool results, loads messages, and creates the next turn.
func (m *Model) startNextTurn(results []domaintools.Result) tea.Cmd {
	m.scope.Info("starting next turn", "result_count", len(results))

	return func() tea.Msg {
		ctx := context.Background()

		// Convert to domain format and persist
		domainResults := make([]domain.ToolResult, len(results))
		for i, r := range results {
			domainResults[i] = domain.ToolResult{
				ToolUseID: r.ToolUseID,
				IsError:   r.IsError(),
				Content:   r.ToMap(),
			}
			if r.Error != nil {
				domainResults[i].Error = r.Error.Message
			}
		}

		msgID, err := m.db.Messages().CreateToolResultMessage(ctx, m.accountID, m.conversationID, domainResults)
		if err != nil {
			m.scope.Error("failed to create tool result message", "error", err)
			return nil
		}

		// Load all messages for the API call
		messages, err := m.db.Messages().List(ctx, m.conversationID)
		if err != nil {
			m.scope.Error("failed to load messages", "error", err)
			return nil
		}

		return nextTurnReady{
			roundID:   m.id,
			messageID: msgID,
			results:   results,
			messages:  messages,
		}
	}
}

// nextTurnReady is an internal message to create the next turn after persistence.
type nextTurnReady struct {
	roundID   domain.MessageID
	messageID domain.MessageID
	results   []domaintools.Result
	messages  []domain.Message
}

// handleNextTurnReady creates and starts the next turn.
func (m *Model) handleNextTurnReady(msg nextTurnReady) tea.Cmd {
	// Create input with tool results (empty text)
	input := msgs.UserSubmittedInput{
		ToolResults: msg.results,
	}

	nextTurn := turn.New(
		m.theme,
		m.conversationID,
		m.accountID,
		msg.messageID,
		input,
		m.width,
		m.db,
		m.chatClient,
		m.toolRegistry,
		m.scope,
	)

	m.turns = append(m.turns, nextTurn)

	return nextTurn.StartStream(msg.messages, nil)
}

// View renders all turns in the round.
func (m *Model) View() string {
	var parts []string

	for i, t := range m.turns {
		// Parent (round) adds gap between children (turns)
		if i > 0 {
			for j := 0; j < gapBetweenTurns; j++ {
				parts = append(parts, "")
			}
		}

		if i == 0 {
			// First turn: show user message + assistant
			parts = append(parts, t.View())
		} else {
			// Subsequent turns: only show assistant response (tool loop continuation)
			parts = append(parts, t.AssistantView())
		}
	}

	// Add divider when round is complete
	if m.state == StateComplete {
		for i := 0; i < gapBeforeDivider; i++ {
			parts = append(parts, "")
		}
		parts = append(parts, m.divider())
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// Height returns the number of lines this component renders.
func (m *Model) Height() int {
	total := 0

	for i, t := range m.turns {
		if i == 0 {
			total += t.Height()
		} else {
			total += lipgloss.Height(t.AssistantView())
		}
		// Add gap between turns (except after last)
		if i < len(m.turns)-1 {
			total += gapBetweenTurns
		}
	}

	// Add divider height when complete
	if m.state == StateComplete {
		total += gapBeforeDivider
		total += 1 // divider line
	}

	return total
}

// divider renders "  ◇ Tero 4s ─────────" (indented to align with assistant text)
func (m *Model) divider() string {
	const indent = 2 // matches assistant paddingWidth

	colors := m.theme.Colors
	muted := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)
	border := lipgloss.NewStyle().Foreground(colors.BorderDefault)

	// Calculate duration
	duration := m.endTime.Sub(m.startTime)
	var durationStr string
	if duration < time.Minute {
		durationStr = fmt.Sprintf("%.1fs", duration.Seconds())
	} else {
		durationStr = fmt.Sprintf("%.1fm", duration.Minutes())
	}

	// Build prefix: "◇ Tero 4s "
	prefix := fmt.Sprintf("◇ Tero %s ", durationStr)
	prefixWidth := lipgloss.Width(prefix)

	// Fill remaining width with ─ (account for indent)
	lineWidth := m.width - indent - prefixWidth
	if lineWidth < 0 {
		lineWidth = 0
	}
	line := strings.Repeat("─", lineWidth)

	return strings.Repeat(" ", indent) + muted.Render(prefix) + border.Render(line)
}

// SetWidth sets the width for all turns.
func (m *Model) SetWidth(width int) {
	m.width = width
	for _, t := range m.turns {
		t.SetWidth(width)
	}
}

// State returns the round's current state.
func (m *Model) State() State {
	return m.state
}

// ID returns the round's ID (first user message ID).
func (m *Model) ID() domain.MessageID {
	return m.id
}
