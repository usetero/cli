package complete

import (
	"os"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// shouldSkipToApp returns true if TERO_SKIP_TO_APP environment variable is set to "true"
// This is a development flag to automatically transition to app mode after onboarding
func shouldSkipToApp() bool {
	return os.Getenv("TERO_SKIP_TO_APP") == "true"
}

// CompleteStep shows the onboarding completion message
type CompleteStep struct {
	// Theme for styling
	theme *styles.Theme

	// Accumulated state from previous steps
	org     api.Organization
	account api.Account

	logger         log.Logger
	width          int
	globalBindings []key.Binding
}

// NewCompleteStep creates a new completion step
func NewCompleteStep(theme *styles.Theme, org api.Organization, account api.Account, logger log.Logger, globalBindings []key.Binding) step.Step {
	if logger == nil {
		panic("logger cannot be nil")
	}

	return &CompleteStep{
		theme:          theme,
		org:            org,
		account:        account,
		logger:         logger,
		width:          80,
		globalBindings: globalBindings,
	}
}

// Init initializes the completion step
func (s *CompleteStep) Init() tea.Cmd {
	s.logger.Info("onboarding complete")
	return nil
}

// Update handles messages
func (s *CompleteStep) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	return s, nil
}

// View renders the completion message
func (s *CompleteStep) View() string {
	themeStyles := s.theme.Styles

	title := themeStyles.Title.Render("You're all set!")

	body1 := themeStyles.Body.Render("We've analyzed your logs and identified waste patterns, quality issues,")
	body2 := themeStyles.Body.Render("and opportunities for improvement.")

	body3 := themeStyles.Body.Render("We're reviewing the results now to make sure everything looks good.")
	body4 := themeStyles.Body.Render("We'll reach out shortly to schedule a walkthrough.")

	contact := themeStyles.Help.Render("Questions in the meantime? Reach out: ") +
		themeStyles.URL.Render("team@usetero.com")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		body1,
		body2,
		"",
		body3,
		body4,
		"",
		contact,
	)
}

// SetSize sets the width available for rendering
func (s *CompleteStep) SetSize(width, height int) {
	s.width = width
}

// IsBusy returns false - no background work
func (s *CompleteStep) IsBusy() bool {
	return false
}

// HasError returns false - no error state
func (s *CompleteStep) HasError() bool {
	return false
}

// Error returns nil - no error
func (s *CompleteStep) Error() error {
	return nil
}

// Next returns (nil, nil) when ready (skip to app enabled), otherwise ErrNotReady
func (s *CompleteStep) Next() (step.Step, error) {
	if shouldSkipToApp() {
		s.logger.Info("skip to app enabled, completing onboarding")
		return nil, nil
	}
	return nil, step.ErrNotReady
}

// Help returns empty key bindings - no actions available
func (s *CompleteStep) Help() help.KeyMap {
	return keymap.Simple{Keys: []key.Binding{}}
}

// Organization returns the organization from completed onboarding
func (s *CompleteStep) Organization() api.Organization {
	return s.org
}

// Account returns the account from completed onboarding
func (s *CompleteStep) Account() api.Account {
	return s.account
}
