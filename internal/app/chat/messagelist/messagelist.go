package messagelist

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/turn"
	"github.com/usetero/cli/internal/app/chat/msgs"
	chatclient "github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

const gap = 2 // blank lines between turns

// Model displays the conversation history and manages turns.
type Model struct {
	theme *styles.Theme
	scope log.Scope

	turns  []*turn.Model
	width  int
	height int

	// Scroll state
	offsetIdx    int  // index of first visible turn
	offsetLine   int  // lines scrolled within that turn
	userScrolled bool // true if user has manually scrolled up (disables auto-scroll)

	// Dependencies
	db           sqlite.DB
	chatClient   chatclient.Client
	toolRegistry *tools.Registry
}

// New creates a new message list.
func New(
	theme *styles.Theme,
	width, height int,
	db sqlite.DB,
	chatClient chatclient.Client,
	toolRegistry *tools.Registry,
	scope log.Scope,
) *Model {
	scope = scope.Child("messagelist")

	return &Model{
		theme:        theme,
		scope:        scope,
		width:        width,
		height:       height,
		db:           db,
		chatClient:   chatClient,
		toolRegistry: toolRegistry,
	}
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.MouseWheelMsg:
		switch msg.Button {
		case tea.MouseWheelUp:
			m.scrollUp(3)
			m.userScrolled = true
		case tea.MouseWheelDown:
			m.scrollDown(3)
			// If user scrolled to bottom, re-enable auto-scroll
			if m.isAtBottom() {
				m.userScrolled = false
			}
		}

	case msgs.TurnStarted:
		// New turn always scrolls to bottom
		m.userScrolled = false
		m.scrollToBottom()

	case msgs.AssistantContentUpdated, msgs.StreamCompleted:
		// Only auto-scroll if user hasn't manually scrolled up
		if !m.userScrolled {
			m.scrollToBottom()
		}
	}

	// Forward to all turns
	for _, t := range m.turns {
		if cmd := t.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}

// StartTurn creates a new turn and begins streaming.
func (m *Model) StartTurn(
	conversationID domain.ConversationID,
	accountID domain.AccountID,
	userMessageID domain.MessageID,
	input msgs.UserSubmittedInput,
	messages []domain.Message,
	context []domain.ContextEntity,
) tea.Cmd {
	m.scope.Debug("starting turn", "user_message_id", userMessageID)

	t := turn.New(
		m.theme,
		conversationID,
		accountID,
		userMessageID,
		input,
		m.width,
		m.db,
		m.chatClient,
		m.toolRegistry,
		m.scope,
	)

	cmd := t.StartStream(messages, context)
	m.turns = append(m.turns, t)

	startCmd := func() tea.Msg {
		return msgs.TurnStarted{
			UserMessageID:  userMessageID,
			ConversationID: conversationID,
		}
	}

	return tea.Batch(cmd, startCmd)
}

// View renders the visible portion of the message list.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 || len(m.turns) == 0 {
		return ""
	}

	var lines []string
	remainingHeight := m.height
	currentIdx := m.offsetIdx
	skipLines := m.offsetLine

	for remainingHeight > 0 && currentIdx < len(m.turns) {
		rendered := m.turns[currentIdx].View()
		turnLines := strings.Split(rendered, "\n")

		// Skip lines for partial scroll
		if skipLines > 0 {
			if skipLines >= len(turnLines) {
				skipLines -= len(turnLines)
				currentIdx++
				continue
			}
			turnLines = turnLines[skipLines:]
			skipLines = 0
		}

		// Add visible lines from this turn
		for _, line := range turnLines {
			if remainingHeight <= 0 {
				break
			}
			lines = append(lines, line)
			remainingHeight--
		}

		// Add gap after turn (if not last and we have room)
		if currentIdx < len(m.turns)-1 {
			for i := 0; i < gap && remainingHeight > 0; i++ {
				lines = append(lines, "")
				remainingHeight--
			}
		}

		currentIdx++
	}

	// Pad to fill height
	for len(lines) < m.height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// SetSize updates the dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	for _, t := range m.turns {
		t.SetWidth(width)
	}
}

// Len returns the number of turns.
func (m *Model) Len() int {
	return len(m.turns)
}

// scrollUp scrolls up by n lines.
func (m *Model) scrollUp(n int) {
	for n > 0 {
		if m.offsetLine > 0 {
			dec := min(n, m.offsetLine)
			m.offsetLine -= dec
			n -= dec
		} else if m.offsetIdx > 0 {
			m.offsetIdx--
			turnHeight := m.turns[m.offsetIdx].Height(m.width)
			m.offsetLine = turnHeight - 1
			if m.offsetIdx < len(m.turns)-1 {
				m.offsetLine += gap
			}
			n--
		} else {
			break
		}
	}
}

// scrollDown scrolls down by n lines, clamping at the bottom.
func (m *Model) scrollDown(n int) {
	maxIdx, maxLine := m.maxScroll()

	for n > 0 {
		// Stop if at max scroll
		if m.offsetIdx > maxIdx || (m.offsetIdx == maxIdx && m.offsetLine >= maxLine) {
			m.offsetIdx = maxIdx
			m.offsetLine = maxLine
			return
		}

		turnHeight := m.turns[m.offsetIdx].Height(m.width)
		remainingInTurn := turnHeight - m.offsetLine - 1
		if m.offsetIdx < len(m.turns)-1 {
			remainingInTurn += gap
		}

		if n <= remainingInTurn {
			m.offsetLine += n
			break
		}

		n -= remainingInTurn + 1
		m.offsetIdx++
		m.offsetLine = 0
	}

	// Clamp to max
	if m.offsetIdx > maxIdx || (m.offsetIdx == maxIdx && m.offsetLine > maxLine) {
		m.offsetIdx = maxIdx
		m.offsetLine = maxLine
	}
}

// maxScroll returns the scroll position where bottom of content meets bottom of viewport.
func (m *Model) maxScroll() (idx int, line int) {
	if len(m.turns) == 0 {
		return 0, 0
	}

	// Calculate total content height
	totalHeight := 0
	for i, t := range m.turns {
		totalHeight += t.Height(m.width)
		if i < len(m.turns)-1 {
			totalHeight += gap
		}
	}

	// Content fits in viewport - no scrolling
	if totalHeight <= m.height {
		return 0, 0
	}

	// Convert max scroll offset to idx/line
	maxOffset := totalHeight - m.height
	for i, t := range m.turns {
		h := t.Height(m.width)
		if i < len(m.turns)-1 {
			h += gap
		}
		if maxOffset < h {
			return i, maxOffset
		}
		maxOffset -= h
	}

	return len(m.turns) - 1, 0
}

// scrollToBottom scrolls to show the last content.
func (m *Model) scrollToBottom() {
	m.offsetIdx, m.offsetLine = m.maxScroll()
}

// isAtBottom returns true if scrolled to the bottom.
func (m *Model) isAtBottom() bool {
	maxIdx, maxLine := m.maxScroll()
	return m.offsetIdx == maxIdx && m.offsetLine == maxLine
}
