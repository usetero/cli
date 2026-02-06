package messagelist

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/round"
	"github.com/usetero/cli/internal/app/chat/msgs"
	chatclient "github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

// roundGap is the number of blank lines between rounds.
const roundGap = 2

// Model displays the conversation history and manages rounds.
// It is a flexible component - it renders exactly the height given by SetSize.
type Model struct {
	theme *styles.Theme
	scope log.Scope

	rounds []*round.Model
	width  int
	height int

	// Scroll state
	scrollOffset int  // lines scrolled from top
	userScrolled bool // true if user has manually scrolled up (disables auto-scroll)

	// Dependencies
	db           sqlite.DB
	chatClient   chatclient.Client
	toolRegistry *tools.Registry
}

// New creates a new message list.
func New(
	theme *styles.Theme,
	db sqlite.DB,
	chatClient chatclient.Client,
	toolRegistry *tools.Registry,
	scope log.Scope,
) *Model {
	scope = scope.Child("messagelist")

	return &Model{
		theme:        theme,
		scope:        scope,
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

	// Forward to all rounds
	for _, r := range m.rounds {
		if cmd := r.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}

// StartTurn creates a new round and begins streaming.
func (m *Model) StartTurn(
	conversationID domain.ConversationID,
	accountID domain.AccountID,
	userMessageID domain.MessageID,
	input msgs.UserSubmittedInput,
	messages []domain.Message,
	context []domain.ContextEntity,
) tea.Cmd {
	m.scope.Debug("starting turn", "user_message_id", userMessageID)

	r := round.New(
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

	cmd := r.StartStream(messages, context)
	m.rounds = append(m.rounds, r)

	startCmd := func() tea.Msg {
		return msgs.TurnStarted{
			UserMessageID:  userMessageID,
			ConversationID: conversationID,
		}
	}

	return tea.Batch(cmd, startCmd)
}

// View renders exactly m.height lines of the message list.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	// Render all content
	content := m.renderContent()
	if content == "" {
		return m.emptyView()
	}

	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	// If content fits, pad to height
	if totalLines <= m.height {
		return m.padToHeight(content)
	}

	// Content exceeds height - apply scroll offset and truncate
	startLine := m.scrollOffset
	if startLine > totalLines-m.height {
		startLine = totalLines - m.height
	}
	if startLine < 0 {
		startLine = 0
	}

	endLine := startLine + m.height
	if endLine > totalLines {
		endLine = totalLines
	}

	visibleLines := lines[startLine:endLine]
	return strings.Join(visibleLines, "\n")
}

// renderContent renders all rounds with gaps between them.
func (m *Model) renderContent() string {
	if len(m.rounds) == 0 {
		return ""
	}

	var parts []string
	for _, r := range m.rounds {
		parts = append(parts, r.View())
	}

	// Join with gaps between rounds
	return strings.Join(parts, "\n\n")
}

// emptyView renders an empty view padded to height.
func (m *Model) emptyView() string {
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render("")
}

// padToHeight pads content to exactly m.height lines.
func (m *Model) padToHeight(content string) string {
	lines := strings.Split(content, "\n")
	for len(lines) < m.height {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

// contentHeight returns the total height of all content.
func (m *Model) contentHeight() int {
	if len(m.rounds) == 0 {
		return 0
	}

	total := 0
	for i, r := range m.rounds {
		total += r.Height()
		if i < len(m.rounds)-1 {
			total += roundGap
		}
	}
	return total
}

// maxScroll returns the maximum scroll offset.
func (m *Model) maxScroll() int {
	contentHeight := m.contentHeight()
	if contentHeight <= m.height {
		return 0
	}
	return contentHeight - m.height
}

// scrollUp scrolls up by n lines.
func (m *Model) scrollUp(n int) {
	m.scrollOffset -= n
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// scrollDown scrolls down by n lines.
func (m *Model) scrollDown(n int) {
	m.scrollOffset += n
	maxOffset := m.maxScroll()
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
}

// scrollToBottom scrolls to show the bottom of content.
func (m *Model) scrollToBottom() {
	m.scrollOffset = m.maxScroll()
}

// isAtBottom returns true if scrolled to the bottom.
func (m *Model) isAtBottom() bool {
	return m.scrollOffset >= m.maxScroll()
}

// SetSize sets the dimensions. This is a flexible component.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	for _, r := range m.rounds {
		r.SetWidth(width)
	}
}

// Len returns the number of rounds.
func (m *Model) Len() int {
	return len(m.rounds)
}
