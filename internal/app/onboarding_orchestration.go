package app

import (
	"time"

	tea "charm.land/bubbletea/v2"

	appevents "github.com/usetero/cli/internal/app/events"
	"github.com/usetero/cli/internal/core/bootstrap"
)

func (m *Model) handleOnboardingMessage(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case bootstrap.OrgSelected:
		return tea.Batch(m.statusBar.Update(msg), m.activateOrg(msg.Org.ID, msg)), true

	case bootstrap.OrgCreated:
		return tea.Batch(m.statusBar.Update(msg), m.activateOrg(msg.Org.ID, msg)), true

	case bootstrap.AccountSelected:
		// Forward to onboarding orchestrator; runtime init happens at EnsureRuntime gate.
		if m.onboarding != nil {
			return m.onboarding.Update(msg), true
		}
		return nil, true

	case bootstrap.EnsureRuntime:
		m.scope.Info("ensuring runtime", "account_id", msg.Account.ID.String())
		start := time.Now()
		catalogCmd, err := m.ensureRuntime(msg.Account.ID.String())
		if err != nil {
			m.scope.Error("failed to ensure runtime", "error", err)
			return appevents.PublishErrorToastCmd("Failed to initialize account runtime", err, true), true
		}
		m.scope.Info("runtime ensured", "account_id", msg.Account.ID.String(), "elapsed_ms", time.Since(start).Milliseconds())
		if m.onboarding != nil {
			return tea.Batch(
				catalogCmd,
				func() tea.Msg { return bootstrap.RuntimeReady(msg) },
			), true
		}
		return catalogCmd, true

	case bootstrap.OnboardingComplete:
		m.state = stateExplorer
		m.user = msg.User
		m.account = msg.Account
		m.scope.Info("onboarding complete",
			"org", msg.Org.Name,
			"account", msg.Account.Name,
		)

		// Create the issue explorer (sizing happens via updateLayout).
		m.explorer = m.newExplorer()
		m.updateLayout()

		return m.explorer.Init(), true
	}

	return nil, false
}
