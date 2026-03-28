package auth

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/ui/browser"
)

var openBrowser = browser.Open

func (m *Model) startDeviceFlow() tea.Cmd {
	return func() tea.Msg {
		flow, err := m.identity.StartDeviceFlow(context.Background())
		if err != nil {
			return deviceFlowFailedMsg{Err: err}
		}
		return deviceFlowStartedMsg{Flow: flow}
	}
}

func (m *Model) pollDeviceFlow() tea.Cmd {
	flow := m.flow
	return func() tea.Msg {
		user, err := m.identity.PollDeviceFlow(context.Background(), flow.DeviceCode, flow.Interval)
		if err != nil {
			return deviceFlowFailedMsg{Err: err}
		}
		return deviceFlowCompletedMsg{User: user}
	}
}

func (m *Model) openBrowser() tea.Cmd {
	url := m.browserURL()
	if url == "" {
		return nil
	}
	return func() tea.Msg {
		return browserOpenedMsg{Err: openBrowser(url)}
	}
}

func (m *Model) browserURL() string {
	if strings.TrimSpace(m.flow.VerificationURIComplete) != "" {
		return strings.TrimSpace(m.flow.VerificationURIComplete)
	}
	return strings.TrimSpace(m.flow.VerificationURI)
}
