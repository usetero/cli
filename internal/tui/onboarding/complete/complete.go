package complete

import (
	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/styles"
)

// CompleteStep shows the onboarding completion message
type CompleteStep struct {
	logger         log.Logger
	width          int
	globalBindings []key.Binding
}

// NewCompleteStep creates a new completion step
func NewCompleteStep(logger log.Logger, globalBindings []key.Binding) step.Step {
	if logger == nil {
		panic("logger cannot be nil")
	}

	return &CompleteStep{
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
	common := styles.Common()

	title := common.Title.Render("You're all set!")

	body1 := common.Body.Render("We've analyzed your logs and identified waste patterns, quality issues,")
	body2 := common.Body.Render("and opportunities for improvement.")

	body3 := common.Body.Render("We're reviewing the results now to make sure everything looks good.")
	body4 := common.Body.Render("We'll reach out shortly to schedule a walkthrough.")

	contact := common.Help.Render("Questions in the meantime? Reach out: ") +
		common.URL.Render("team@usetero.com")

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

// IsComplete returns false - this step stays visible
func (s *CompleteStep) IsComplete() bool {
	return false
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

// Next returns nil - this is the final step
func (s *CompleteStep) Next() step.Step {
	return nil
}

// Help returns empty key bindings - no actions available
func (s *CompleteStep) Help() help.KeyMap {
	return keymap.Simple{Keys: []key.Binding{}}
}
