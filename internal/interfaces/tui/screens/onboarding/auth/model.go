package auth

import (
	"context"
	"errors"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/identity"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

type phase uint8

const (
	phaseIdle phase = iota
	phaseStarting
	phaseWaiting
	phaseAuthenticated
	phaseFailed
)

// Model owns the onboarding device-auth step.
type Model struct {
	scope    logging.Scope
	identity *identity.Service
	theme    theme.Theme
	width    int
	height   int
	phase    phase
	flow     identity.DeviceFlow
	user     identity.User
	err      error
}

var _ core.Model = (*Model)(nil)
var _ core.BusyProvider = (*Model)(nil)
var _ core.InputProvider = (*Model)(nil)

func New(scope logging.Scope, identityService *identity.Service, appTheme theme.Theme) *Model {
	if identityService == nil {
		panic("auth model requires identity service")
	}
	return &Model{
		scope:    scope,
		identity: identityService,
		theme:    appTheme,
		phase:    phaseIdle,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.KeyPressMsg:
		if m.phase == phaseIdle && key.Matches(typed, startBinding) {
			m.phase = phaseStarting
			m.err = nil
			return m, m.startDeviceFlow()
		}
		if m.phase == phaseFailed && key.Matches(typed, retryBinding) {
			if m.browserURL() == "" {
				m.phase = phaseStarting
				m.err = nil
				return m, m.startDeviceFlow()
			}
		}
		if (m.phase == phaseWaiting || m.phase == phaseFailed) && key.Matches(typed, reopenBinding) {
			return m, m.openBrowser()
		}
	case deviceFlowStartedMsg:
		m.phase = phaseWaiting
		m.flow = typed.Flow
		m.err = nil
		return m, tea.Batch(m.openBrowser(), m.pollDeviceFlow())
	case browserOpenedMsg:
		if typed.Err != nil {
			m.scope.Warn("open auth browser", "error", typed.Err)
		}
		return m, nil
	case deviceFlowCompletedMsg:
		m.phase = phaseAuthenticated
		m.user = typed.User
		m.err = nil
		m.scope.Info("device auth completed", "user_id", typed.User.ID, "email", typed.User.Email)
		return m, nil
	case deviceFlowFailedMsg:
		m.phase = phaseFailed
		m.err = typed.Err
		if typed.Err != nil && !errors.Is(typed.Err, context.Canceled) {
			m.scope.Error("device auth failed", "error", typed.Err)
		}
		return m, nil
	}

	return m, nil
}

// Authenticated reports whether the device-auth step completed successfully.
func (m *Model) Authenticated() bool { return m.phase == phaseAuthenticated }
