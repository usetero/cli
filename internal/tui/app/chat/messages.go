package chat

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/sqlite/gen"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/spinner"
	"github.com/usetero/cli/internal/upload"
)

// messagesLoadedMsg is sent when messages are loaded from SQLite.
type messagesLoadedMsg struct {
	messages []gen.Message
	err      error
}

// tablesChangedMsg is sent when SQLite tables change.
type tablesChangedMsg struct {
	tables []string
}

// UploadEventMsg wraps upload events for the Bubble Tea message loop.
type UploadEventMsg struct {
	Event upload.Event
}

// Messages provides message data and state for the message list.
type Messages interface {
	Init() tea.Cmd
	Update(msg tea.Msg) tea.Cmd
	SetConversation(conversationID string) tea.Cmd
	Refresh() tea.Cmd
	SetWidth(width int)
	Items() []Item
	HasError() bool
	Error() error
	IsBusy() bool
	Close() error
}

// Compile-time check that sqliteMessages implements Messages.
var _ Messages = (*sqliteMessages)(nil)

// sqliteMessages loads messages from SQLite.
type sqliteMessages struct {
	theme *styles.Theme
	db    sqlite.Database

	conversationID string
	items          []Item
	itemMap        map[string]Item
	width          int

	subscription *sqlite.Subscription
	err          error
}

// NewMessages creates a new messages data manager.
func NewMessages(theme *styles.Theme, db sqlite.Database) Messages {
	return &sqliteMessages{
		theme:   theme,
		db:      db,
		itemMap: make(map[string]Item),
	}
}

// Init starts listening for database changes.
func (m *sqliteMessages) Init() tea.Cmd {
	m.subscription = m.db.Subscribe()
	return m.listenForChanges()
}

// Update handles data-related messages.
func (m *sqliteMessages) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case messagesLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return nil
		}
		m.err = nil
		return m.buildItems(msg.messages)

	case tablesChangedMsg:
		for _, table := range msg.tables {
			if table == "messages" {
				return tea.Batch(m.Refresh(), m.listenForChanges())
			}
		}
		return m.listenForChanges()

	case UploadEventMsg:
		return m.handleUploadEvent(msg.Event)

	case spinner.TickMsg:
		// Route tick messages to spinning items
		for _, item := range m.items {
			if item.Spinning() {
				if cmd := item.Update(msg); cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
		return tea.Batch(cmds...)
	}

	return nil
}

// SetConversation sets the current conversation and loads messages.
func (m *sqliteMessages) SetConversation(conversationID string) tea.Cmd {
	m.conversationID = conversationID
	m.items = nil
	m.itemMap = make(map[string]Item)
	m.err = nil

	if conversationID == "" || m.db == nil {
		return nil
	}

	return m.load()
}

// Refresh reloads messages from SQLite.
func (m *sqliteMessages) Refresh() tea.Cmd {
	if m.conversationID == "" || m.db == nil {
		return nil
	}
	return m.load()
}

// SetWidth sets the width for item rendering.
func (m *sqliteMessages) SetWidth(width int) {
	m.width = width
	for _, item := range m.items {
		item.SetWidth(width)
	}
}

// Items returns the current list of items.
func (m *sqliteMessages) Items() []Item {
	return m.items
}

// HasError returns true if there's an error.
func (m *sqliteMessages) HasError() bool {
	return m.err != nil
}

// Error returns the current error.
func (m *sqliteMessages) Error() error {
	return m.err
}

// IsBusy returns true if any item is spinning.
func (m *sqliteMessages) IsBusy() bool {
	for _, item := range m.items {
		if item.Spinning() {
			return true
		}
	}
	return false
}

// Close releases resources.
func (m *sqliteMessages) Close() error {
	if m.subscription != nil {
		m.subscription.Stop()
		m.subscription = nil
	}
	return nil
}

// load loads messages from SQLite.
func (m *sqliteMessages) load() tea.Cmd {
	return func() tea.Msg {
		messages, err := m.db.Messages().List(context.Background(), m.conversationID)
		return messagesLoadedMsg{messages: messages, err: err}
	}
}

// listenForChanges waits for database changes.
func (m *sqliteMessages) listenForChanges() tea.Cmd {
	if m.subscription == nil {
		return nil
	}
	return func() tea.Msg {
		tables, ok := <-m.subscription.Changes()
		if !ok {
			return nil
		}
		return tablesChangedMsg{tables: tables}
	}
}

// handleUploadEvent processes upload events.
func (m *sqliteMessages) handleUploadEvent(event upload.Event) tea.Cmd {
	switch e := event.(type) {
	case upload.MessageProcessingEvent:
		if e.ConversationID != m.conversationID {
			return nil
		}

		assistant := NewAssistantMessage(m.theme)
		assistant.SetWidth(m.width)
		m.items = append(m.items, assistant)
		return assistant.Init()
	}
	return nil
}

// buildItems converts SQLite messages into Items.
func (m *sqliteMessages) buildItems(messages []gen.Message) tea.Cmd {
	var cmds []tea.Cmd
	var items []Item
	itemMap := make(map[string]Item)

	// Find pending assistant to preserve it
	var pending *AssistantMessage
	for _, item := range m.items {
		if am, ok := item.(*AssistantMessage); ok && am.ID() == "" {
			pending = am
			break
		}
	}

	for _, msg := range messages {
		if msg.Role == nil || msg.ID == nil {
			continue
		}

		role := chat.MessageRole(*msg.Role)
		id := *msg.ID

		switch role {
		case chat.RoleUser:
			item := m.getOrCreateUserMessage(id)
			item.SetWidth(m.width)
			if msg.Content != nil {
				if blocks, err := block.Parse(*msg.Content); err == nil {
					item.SetContent(blocks)
				}
			}
			items = append(items, item)
			itemMap[id] = item

		case chat.RoleAssistant:
			var item *AssistantMessage
			if pending != nil {
				pending.SetMessageID(id)
				item = pending
				pending = nil
			} else {
				item = m.getOrCreateAssistantMessage(id)
			}

			item.SetWidth(m.width)
			if msg.Content != nil {
				if blocks, err := block.Parse(*msg.Content); err == nil {
					item.SetContent(blocks)
				}
			}
			items = append(items, item)
			itemMap[id] = item

			if cmd := item.Init(); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	if pending != nil {
		items = append(items, pending)
	}

	m.items = items
	m.itemMap = itemMap

	return tea.Batch(cmds...)
}

func (m *sqliteMessages) getOrCreateUserMessage(id string) *UserMessage {
	if existing, ok := m.itemMap[id]; ok {
		if msg, ok := existing.(*UserMessage); ok {
			return msg
		}
	}
	return NewUserMessage(m.theme, id)
}

func (m *sqliteMessages) getOrCreateAssistantMessage(id string) *AssistantMessage {
	if existing, ok := m.itemMap[id]; ok {
		if msg, ok := existing.(*AssistantMessage); ok {
			return msg
		}
	}
	return NewAssistantMessageWithID(m.theme, id)
}
