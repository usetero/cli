package app

import (
	"context"
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync/powersynctest"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/page"
	"github.com/usetero/cli/internal/tui/components/commandbar"
)

func TestApp_New(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	theme := styles.NewTheme(true)
	db := powersynctest.OpenTestDB(t)
	org := apitest.NewOrganization()
	account := apitest.NewAccount()
	workspace := apitest.NewWorkspace()
	logger := logtest.New(t)

	app := New(ctx, theme, db, org, account, workspace, logger, nil)

	if app == nil {
		t.Fatal("expected app to be created")
	}
	if app.chat == nil {
		t.Error("expected chat to be initialized")
	}
	if app.commandBar == nil {
		t.Error("expected command bar to be initialized")
	}
	if app.sidebar == nil {
		t.Error("expected sidebar to be initialized")
	}
	if app.header == nil {
		t.Error("expected header to be initialized")
	}
}

func TestApp_SetSize(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	theme := styles.NewTheme(true)
	db := powersynctest.OpenTestDB(t)
	org := apitest.NewOrganization()
	account := apitest.NewAccount()
	workspace := apitest.NewWorkspace()

	t.Run("wide mode when width >= threshold", func(t *testing.T) {
		t.Parallel()

		logger := logtest.New(t)
		app := New(ctx, theme, db, org, account, workspace, logger, nil)
		app.SetSize(CompactModeWidth+1, 40)

		if app.compact {
			t.Error("expected wide mode (compact=false) for width > threshold")
		}
	})

	t.Run("compact mode when width < threshold", func(t *testing.T) {
		t.Parallel()

		logger := logtest.New(t)
		app := New(ctx, theme, db, org, account, workspace, logger, nil)
		app.SetSize(CompactModeWidth-1, 40)

		if !app.compact {
			t.Error("expected compact mode (compact=true) for width < threshold")
		}
	})
}

func TestApp_View(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	theme := styles.NewTheme(true)
	db := powersynctest.OpenTestDB(t)
	org := apitest.NewOrganization()
	account := apitest.NewAccount()
	workspace := apitest.NewWorkspace()

	t.Run("returns empty when size not set", func(t *testing.T) {
		t.Parallel()

		logger := logtest.New(t)
		app := New(ctx, theme, db, org, account, workspace, logger, nil)
		result := app.View()

		if result != "" {
			t.Errorf("expected empty string before SetSize, got %q", result)
		}
	})

	t.Run("renders content after SetSize", func(t *testing.T) {
		t.Parallel()

		logger := logtest.New(t)
		app := New(ctx, theme, db, org, account, workspace, logger, nil)
		app.SetSize(150, 40)

		result := app.View()

		if result == "" {
			t.Error("expected non-empty view after SetSize")
		}
	})

	t.Run("includes sidebar in wide mode", func(t *testing.T) {
		t.Parallel()

		logger := logtest.New(t)
		app := New(ctx, theme, db, org, account, workspace, logger, nil)
		app.SetSize(150, 40)

		result := app.View()

		// Sidebar should show org name
		if !strings.Contains(result, org.Name) {
			t.Errorf("expected org name %q in sidebar", org.Name)
		}
	})
}

func TestApp_FocusStack(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	theme := styles.NewTheme(true)
	db := powersynctest.OpenTestDB(t)
	org := apitest.NewOrganization()
	account := apitest.NewAccount()
	workspace := apitest.NewWorkspace()

	t.Run("focused page is chat by default", func(t *testing.T) {
		t.Parallel()

		logger := logtest.New(t)
		app := New(ctx, theme, db, org, account, workspace, logger, nil)

		focused := app.focusedPage()

		if focused != app.chat {
			t.Error("expected chat to be focused by default")
		}
	})

	t.Run("push adds page to stack", func(t *testing.T) {
		t.Parallel()

		logger := logtest.New(t)
		app := New(ctx, theme, db, org, account, workspace, logger, nil)
		mockPage := &mockPage{title: "Mock Page"}

		app.PushFocus(mockPage)

		if len(app.focusStack) != 1 {
			t.Errorf("expected 1 page on stack, got %d", len(app.focusStack))
		}
		if app.focusedPage().Title() != mockPage.Title() {
			t.Error("expected mock page to be focused")
		}
	})

	t.Run("pop removes page from stack", func(t *testing.T) {
		t.Parallel()

		logger := logtest.New(t)
		app := New(ctx, theme, db, org, account, workspace, logger, nil)
		mockPage := &mockPage{title: "Mock Page"}

		app.PushFocus(mockPage)
		app.PopFocus()

		if len(app.focusStack) != 0 {
			t.Errorf("expected empty stack after pop, got %d", len(app.focusStack))
		}
		if app.focusedPage() != app.chat {
			t.Error("expected chat to be focused after pop")
		}
	})

	t.Run("escape dismisses focus stack", func(t *testing.T) {
		t.Parallel()

		logger := logtest.New(t)
		app := New(ctx, theme, db, org, account, workspace, logger, nil)
		mockPage := &mockPage{title: "Mock Page"}

		app.PushFocus(mockPage)
		app.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

		if len(app.focusStack) != 0 {
			t.Error("expected escape to dismiss focus stack")
		}
	})
}

func TestApp_SendMessage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	theme := styles.NewTheme(true)
	org := apitest.NewOrganization()
	account := apitest.NewAccount()
	workspace := apitest.NewWorkspace()

	t.Run("creates conversation on first message", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		logger := logtest.New(t)
		app := New(ctx, theme, db, org, account, workspace, logger, nil)
		app.SetSize(150, 40)

		// Submit a message - this returns a command that creates the conversation
		cmd := app.Update(commandbar.SubmitMsg{Text: "Hello world"})

		// Execute the command to get the result
		msg := cmd()
		sentMsg, ok := msg.(messageSentMsg)
		if !ok {
			t.Fatalf("expected messageSentMsg, got %T", msg)
		}

		if sentMsg.err != nil {
			t.Fatalf("unexpected error: %v", sentMsg.err)
		}
		if sentMsg.conversationID == "" {
			t.Error("expected conversation ID to be set")
		}

		// Verify conversation exists in database
		accountID := account.ID
		convs, err := db.Queries().ListConversationsByAccount(ctx, &accountID)
		if err != nil {
			t.Fatalf("failed to list conversations: %v", err)
		}
		if len(convs) != 1 {
			t.Errorf("expected 1 conversation, got %d", len(convs))
		}
	})

	t.Run("reuses conversation for subsequent messages", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		logger := logtest.New(t)
		app := New(ctx, theme, db, org, account, workspace, logger, nil)
		app.SetSize(150, 40)

		// First message - creates conversation
		cmd := app.Update(commandbar.SubmitMsg{Text: "First"})
		msg := cmd()
		firstSent, ok := msg.(messageSentMsg)
		if !ok {
			t.Fatalf("expected messageSentMsg, got %T", msg)
		}
		firstConvID := firstSent.conversationID

		// Process the message to set conversationID on app
		app.Update(firstSent)

		// Second message - reuses conversation
		cmd = app.Update(commandbar.SubmitMsg{Text: "Second"})
		msg = cmd()
		secondSent, ok := msg.(messageSentMsg)
		if !ok {
			t.Fatalf("expected messageSentMsg, got %T", msg)
		}

		// Second message should reuse the conversation
		// (conversationID is empty in the result because it was already set on app)
		if secondSent.conversationID != "" && secondSent.conversationID != firstConvID {
			t.Errorf("expected same or empty conversation ID, got first=%s second=%s", firstConvID, secondSent.conversationID)
		}

		// Verify only 1 conversation exists
		accountID := account.ID
		convs, err := db.Queries().ListConversationsByAccount(ctx, &accountID)
		if err != nil {
			t.Fatalf("failed to list conversations: %v", err)
		}
		if len(convs) != 1 {
			t.Errorf("expected 1 conversation, got %d", len(convs))
		}
	})
}

func TestApp_State(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	theme := styles.NewTheme(true)
	db := powersynctest.OpenTestDB(t)
	org := apitest.NewOrganization()
	account := apitest.NewAccount()
	workspace := apitest.NewWorkspace()

	t.Run("IsComplete always returns false", func(t *testing.T) {
		t.Parallel()

		logger := logtest.New(t)
		app := New(ctx, theme, db, org, account, workspace, logger, nil)

		if app.IsComplete() {
			t.Error("expected IsComplete to return false")
		}
	})

	t.Run("IsBusy reflects chat state", func(t *testing.T) {
		t.Parallel()

		logger := logtest.New(t)
		app := New(ctx, theme, db, org, account, workspace, logger, nil)

		// Chat is not busy by default
		if app.IsBusy() {
			t.Error("expected IsBusy to return false initially")
		}
	})

	t.Run("HasError reflects focused page state", func(t *testing.T) {
		t.Parallel()

		logger := logtest.New(t)
		app := New(ctx, theme, db, org, account, workspace, logger, nil)

		// No error by default
		if app.HasError() {
			t.Error("expected HasError to return false initially")
		}
	})
}

func TestApp_GlobalBindings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	theme := styles.NewTheme(true)
	db := powersynctest.OpenTestDB(t)
	org := apitest.NewOrganization()
	account := apitest.NewAccount()
	workspace := apitest.NewWorkspace()
	logger := logtest.New(t)

	bindings := []key.Binding{
		key.NewBinding(key.WithKeys("ctrl+q"), key.WithHelp("ctrl+q", "quit")),
	}

	app := New(ctx, theme, db, org, account, workspace, logger, bindings)
	app.SetSize(150, 40)

	// Global bindings should be stored
	if len(app.globalBindings) != 1 {
		t.Errorf("expected 1 global binding, got %d", len(app.globalBindings))
	}
}

// mockPage implements page.Page for testing
type mockPage struct {
	title string
	busy  bool
	err   error
}

func (m *mockPage) Init() tea.Cmd                { return nil }
func (m *mockPage) Update(tea.Msg) tea.Cmd       { return nil }
func (m *mockPage) View() string                 { return m.title }
func (m *mockPage) SetSize(int, int)             {}
func (m *mockPage) Title() string                { return m.title }
func (m *mockPage) Metadata() []page.Metadata    { return nil }
func (m *mockPage) AcceptsNaturalLanguage() bool { return false }
func (m *mockPage) Commands() []page.Command     { return nil }
func (m *mockPage) KeyBindings() []key.Binding   { return nil }
func (m *mockPage) IsBusy() bool                 { return m.busy }
func (m *mockPage) HasError() bool               { return m.err != nil }
func (m *mockPage) Error() error                 { return m.err }
