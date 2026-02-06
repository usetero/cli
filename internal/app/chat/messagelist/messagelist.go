package messagelist

import (
	"strings"

	"charm.land/bubbles/v2/key"
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

var (
	scrollUpKey   = key.NewBinding(key.WithKeys("up"))
	scrollDownKey = key.NewBinding(key.WithKeys("down"))
)

// roundGap is the number of blank lines between rounds.
const roundGap = 2

// focusBorderWidth is the width consumed by the left border when focused.
const focusBorderWidth = 1

// Model displays the conversation history and manages rounds.
// It is a flexible component - it renders exactly the height given by SetSize.
type Model struct {
	theme *styles.Theme
	scope log.Scope

	rounds []*round.Model
	width  int
	height int

	// Focus state
	focused bool // true when message list has keyboard focus

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
	case tea.KeyPressMsg:
		if m.focused {
			if key.Matches(msg, scrollUpKey) {
				m.scrollUp(3)
				m.userScrolled = true
			} else if key.Matches(msg, scrollDownKey) {
				m.scrollDown(3)
				if m.isAtBottom() {
					m.userScrolled = false
				}
			}
		}

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
		m.contentWidth(),
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
		lines = m.padLines(lines)
	} else {
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

		lines = lines[startLine:endLine]
	}

	output := strings.Join(lines, "\n")

	borderColor := m.theme.Colors.Page.Bg
	if m.focused {
		borderColor = m.theme.Colors.AccentAlt
	}

	return lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(borderColor).
		Render(output)
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

// padLines pads lines to exactly m.height.
func (m *Model) padLines(lines []string) []string {
	for len(lines) < m.height {
		lines = append(lines, "")
	}
	return lines
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
	m.updateRoundWidths()
}

// SetFocused sets focus state.
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

// contentWidth returns the width available for round content.
// Border is always present (invisible when unfocused) so always subtract.
func (m *Model) contentWidth() int {
	return m.width - focusBorderWidth
}

// updateRoundWidths sets the width on all rounds.
func (m *Model) updateRoundWidths() {
	w := m.contentWidth()
	for _, r := range m.rounds {
		r.SetWidth(w)
	}
}

// Focused returns whether the message list is focused.
func (m *Model) Focused() bool {
	return m.focused
}

// Len returns the number of rounds.
func (m *Model) Len() int {
	return len(m.rounds)
}
