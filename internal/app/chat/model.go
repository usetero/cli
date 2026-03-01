package chat

import (
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/chat/inputbar"
	"github.com/usetero/cli/internal/app/chat/messagelist"
	"github.com/usetero/cli/internal/app/chat/usecase"
	"github.com/usetero/cli/internal/app/chattools"
	"github.com/usetero/cli/internal/auth"
	corechat "github.com/usetero/cli/internal/core/chat"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

const dbOpTimeout = 2 * time.Second

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
	session        *corechat.Session

	user      *auth.User
	account   domain.Account
	workspace domain.Workspace
	theme     styles.Theme
	width     int
	height    int
	originX   int
	originY   int

	// Empty state
	policySummary *domain.AccountSummary

	// Dependencies
	db           sqlite.DB
	runtimeDeps  usecase.RuntimeDeps
	toolRegistry *tools.Registry
}

// emptyStatePollTickMsg triggers a policy summary fetch for the empty state.
type emptyStatePollTickMsg struct{}

// emptyStateSummaryLoadedMsg carries an async empty-state summary fetch result.
type emptyStateSummaryLoadedMsg struct {
	summary domain.AccountSummary
	err     error
}

// New creates a new chat model.
func New(
	user *auth.User,
	account domain.Account,
	workspace domain.Workspace,
	theme styles.Theme,
	db sqlite.DB,
	runtimeDeps usecase.RuntimeDeps,
	toolRegistry *tools.Registry,
	scope log.Scope,
) *Model {
	scope = scope.Child("chat")

	return &Model{
		scope:        scope,
		inputBar:     inputbar.New(user, theme, scope),
		messageList:  messagelist.New(theme, runtimeDeps, toolRegistry, scope),
		user:         user,
		account:      account,
		workspace:    workspace,
		theme:        theme,
		db:           db,
		runtimeDeps:  runtimeDeps,
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
		return emptyStatePollTickMsg{}
	})
}
