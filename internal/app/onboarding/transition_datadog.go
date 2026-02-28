package onboarding

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/onboarding/msgs"
)

func (m *Model) handleDatadogTransition(msg tea.Msg) (TransitionOutcome, bool) {
	switch msg := msg.(type) {
	case msgs.DatadogReady:
		m.scope.Debug("datadog ready")
		return advance(GateWorkspaceSelect), true

	case msgs.DatadogNeeded:
		m.scope.Debug("datadog setup needed")
		return advance(GateDatadogRegion), true

	case msgs.DatadogRegionSelected:
		m.scope.Info("datadog region selected", slog.String("site", string(msg.Site)))
		m.state.ddSite = msg.Site
		return advance(GateDatadogAPIKey), true

	case msgs.DatadogAPIKeyEntered:
		m.scope.Debug("datadog api key validated")
		m.state.ddAPIKey = msg.APIKey
		return advance(GateDatadogAppKey), true

	case msgs.DatadogAccountCreated:
		m.scope.Info("datadog account created", slog.String("datadog_account_id", msg.DatadogAccountID.String()))
		m.state.ddAccount = msg.DatadogAccountID
		return advance(GateDatadogDiscovery), true

	case msgs.DatadogDiscoveryComplete:
		m.scope.Info("datadog discovery complete")
		return advance(GateWorkspaceSelect), true

	default:
		return noop(), false
	}
}
