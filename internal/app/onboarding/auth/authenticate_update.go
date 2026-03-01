package auth

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
	"github.com/pkg/browser"

	appevents "github.com/usetero/cli/internal/app/events"
	"github.com/usetero/cli/internal/core/bootstrap"
)

// Update handles messages.
func (m *AuthenticateModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case deviceAuthMsg:
		return m.handleDeviceAuth(msg)
	case authCompleteMsg:
		return m.handleAuthComplete(msg)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd
	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}
	return nil
}

func (m *AuthenticateModel) handleDeviceAuth(msg deviceAuthMsg) tea.Cmd {
	if msg.err != nil {
		m.scope.Error("device auth failed", "error", msg.err)
		m.err = msg.err
		return appevents.PublishErrorToastCmd("Failed to start authentication", msg.err, false)
	}

	m.device = msg.deviceAuth
	m.state = stateReady
	m.scope.Debug("device code received", "code", m.device.UserCode)

	if err := browser.OpenURL(m.device.VerificationURIComplete); err != nil {
		m.scope.Warn("failed to open browser", "error", err)
		m.browserFailed = true
		return nil
	}
	m.scope.Debug("browser opened")
	m.state = statePolling
	return m.pollForAuth()
}

func (m *AuthenticateModel) handleAuthComplete(msg authCompleteMsg) tea.Cmd {
	if msg.err != nil {
		m.scope.Error("authentication failed", "error", msg.err)
		m.err = msg.err
		m.state = stateReady
		return appevents.PublishErrorToastCmd("Authentication failed", msg.err, false)
	}

	m.state = stateComplete
	m.scope.Info("user authenticated", "user_id", msg.result.User.ID)
	return func() tea.Msg {
		return bootstrap.Authenticated{User: msg.result.User}
	}
}

func (m *AuthenticateModel) handleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	if m.state != stateReady {
		return nil
	}

	switch msg.String() {
	case "enter":
		if err := browser.OpenURL(m.device.VerificationURIComplete); err != nil {
			m.scope.Warn("failed to open browser", "error", err)
			m.browserFailed = true
			return nil
		}
		m.scope.Debug("browser opened")
		m.browserFailed = false
		m.state = statePolling
		return m.pollForAuth()

	case "c":
		if err := clipboard.WriteAll(m.device.VerificationURIComplete); err != nil {
			m.scope.Warn("failed to copy to clipboard", "error", err)
			return nil
		}
		m.scope.Debug("url copied to clipboard")
		m.copiedToClipboard = true
		m.state = statePolling
		return m.pollForAuth()

	case "r":
		if m.err != nil {
			m.scope.Debug("retrying authentication")
			m.err = nil
			m.state = stateInitializing
			return m.Init()
		}
	}
	return nil
}
