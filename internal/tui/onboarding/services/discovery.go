package services

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/spinner"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/log_events"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/styles"
)

const (
	pollInterval = 2 * time.Second // How often to check discovery status
)

// discoveryStatusMsg is sent when we receive discovery status from the control plane
type discoveryStatusMsg struct {
	completed      bool
	serviceCount   int
	discoveryError string // Last error from discovery job (recoverable)
	err            error  // API/system error
}

// ServiceDiscoveryPoller polls for service discovery status
type ServiceDiscoveryPoller interface {
	GetServiceDiscoveryStatus(ctx context.Context, datadogAccountID string) (*api.ServiceDiscoveryStatus, error)
}

// DiscoveryStep waits for service discovery to complete
type DiscoveryStep struct {
	// Accumulated state from previous steps
	role             string
	orgID            string
	accountID        string
	datadogAccountID *string

	// Services (defined by consumer interfaces)
	discoveryPoller ServiceDiscoveryPoller

	// Pass-through to next step
	apiClient api.Client
	logger    log.Logger

	// UI state
	spinner        spinner.Model
	checking       bool
	complete       bool
	serviceCount   int
	discoveryError string // Last error from discovery job (recoverable)
	err            error  // API/system error
	width          int
	globalBindings []key.Binding
}

// NewDiscoveryStep creates a new service discovery step
func NewDiscoveryStep(
	role string,
	orgID string,
	accountID string,
	datadogAccountID *string,
	discoveryPoller ServiceDiscoveryPoller,
	apiClient api.Client,
	logger log.Logger,
	globalBindings []key.Binding,
) step.Step {
	if discoveryPoller == nil {
		panic("discoveryPoller cannot be nil")
	}
	if apiClient == nil {
		panic("apiClient cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	theme := styles.CurrentTheme()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	return &DiscoveryStep{
		role:             role,
		orgID:            orgID,
		accountID:        accountID,
		datadogAccountID: datadogAccountID,
		discoveryPoller:  discoveryPoller,
		apiClient:        apiClient,
		logger:           logger,
		spinner:          s,
		width:            80,
		globalBindings:   globalBindings,
	}
}

// Init starts the spinner and begins polling for discovery status
func (s *DiscoveryStep) Init() tea.Cmd {
	return tea.Batch(
		s.spinner.Tick,
		s.checkDiscoveryStatus(),
	)
}

// Update handles messages
func (s *DiscoveryStep) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	switch msg := msg.(type) {
	case discoveryStatusMsg:
		s.checking = false
		if msg.err != nil {
			s.logger.Error("failed to check discovery status", "error", msg.err)
			s.err = msg.err
			return s, nil
		}

		// Clear any previous error
		s.err = nil

		// Update service count and discovery error
		s.serviceCount = msg.serviceCount
		s.discoveryError = msg.discoveryError

		// Log discovery errors (but keep polling)
		if s.discoveryError != "" {
			s.logger.Warn("service discovery has errors",
				log.String("error", s.discoveryError))
		}

		if msg.completed {
			s.logger.Info("service discovery completed", log.Int("count", s.serviceCount))
			s.complete = true
			return s, nil
		}

		// Not complete yet, poll again after interval
		return s, tea.Tick(pollInterval, func(t time.Time) tea.Msg {
			return s.checkDiscoveryStatus()()
		})

	case spinner.TickMsg:
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd

	case tea.KeyMsg:
		// Allow retry on error
		if s.err != nil && msg.String() == "enter" {
			s.err = nil
			s.checking = true
			return s, s.checkDiscoveryStatus()
		}
	}

	return s, nil
}

// checkDiscoveryStatus polls the control plane for service discovery completion
func (s *DiscoveryStep) checkDiscoveryStatus() tea.Cmd {
	return func() tea.Msg {
		if s.datadogAccountID == nil {
			return discoveryStatusMsg{err: fmt.Errorf("no datadog account specified")}
		}

		s.logger.Debug("checking service discovery status",
			log.String("datadogAccountID", *s.datadogAccountID))

		ctx := context.Background()

		// Query service discovery status from the DatadogAccount
		status, err := s.discoveryPoller.GetServiceDiscoveryStatus(ctx, *s.datadogAccountID)
		if err != nil {
			return discoveryStatusMsg{err: err}
		}

		if status == nil {
			s.logger.Debug("no Datadog account found")
			return discoveryStatusMsg{err: fmt.Errorf("datadog account not found")}
		}

		s.logger.Debug("service discovery status",
			log.String("status", string(status.Status)),
			log.Int("servicesDiscovered", status.ServicesDiscovered))

		// Check if discovery is ready
		completed := status.Status == api.DiscoveryStatusReady

		return discoveryStatusMsg{
			completed:      completed,
			serviceCount:   status.ServicesDiscovered,
			discoveryError: status.LastError,
		}
	}
}

// View renders the service discovery UI
func (s *DiscoveryStep) View() string {
	common := styles.Common()

	title := common.Title.Render("Discovering services...")
	subtitle := common.Subtitle.Render("This usually takes 30-60 seconds.")

	// Build status message with spinner and count
	var statusMsg string
	if s.serviceCount > 0 {
		statusMsg = s.spinner.View() + " " + common.Body.Render(fmt.Sprintf("%d services discovered so far", s.serviceCount))
	} else {
		statusMsg = s.spinner.View() + " " + common.Body.Render("Starting discovery...")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		subtitle,
		"",
		statusMsg,
	)
}

// SetSize sets the width available for rendering
func (s *DiscoveryStep) SetSize(width, height int) {
	s.width = width
}

// IsComplete returns true if service discovery has completed successfully
func (s *DiscoveryStep) IsComplete() bool {
	return s.complete && s.err == nil
}

// IsBusy returns true while checking discovery status
func (s *DiscoveryStep) IsBusy() bool {
	// Not busy if there's an error or if complete
	if s.err != nil || s.complete {
		return false
	}
	return s.checking || !s.complete
}

// HasError returns true if there's a discovery error or API error
func (s *DiscoveryStep) HasError() bool {
	return s.err != nil || s.discoveryError != ""
}

// Error returns the current error (discovery or API error)
func (s *DiscoveryStep) Error() error {
	if s.err != nil {
		return s.err
	}
	if s.discoveryError != "" {
		return fmt.Errorf("%s", s.discoveryError)
	}
	return nil
}

// Next returns the log event discovery step
func (s *DiscoveryStep) Next() step.Step {
	// Control plane auto-enables top services, so we go directly to log event discovery
	return log_events.NewDiscoveryStep(s.role, s.orgID, s.accountID, s.datadogAccountID, s.apiClient, s.logger, s.globalBindings)
}

// Help returns the key bindings for this step
func (s *DiscoveryStep) Help() help.KeyMap {
	// Show retry option if there's an error
	if s.err != nil {
		return keymap.Simple{
			Keys: []key.Binding{
				key.NewBinding(
					key.WithKeys("enter"),
					key.WithHelp("enter", "retry"),
				),
			},
		}
	}

	// No user interaction during normal polling
	return keymap.Simple{
		Keys: []key.Binding{},
	}
}
