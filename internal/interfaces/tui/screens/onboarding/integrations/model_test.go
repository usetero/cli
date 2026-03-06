package integrationsflow

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/infrastructure/logging"
	datadogapikey "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations/datadog/api_key"
	datadogappkey "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations/datadog/app_key"
	datadogregion "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations/datadog/region"
	providerselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations/provider/select"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
	"github.com/usetero/cli/internal/runtime/onboarding"
)

func newModel() *Model {
	appTheme := theme.New(false)
	return New(
		providerselect.New(logging.Scope{}, appTheme),
		datadogregion.New(logging.Scope{}, appTheme),
		datadogapikey.New(logging.Scope{}, appTheme),
		datadogappkey.New(logging.Scope{}, appTheme),
		[]integrations.Provider{
			integrations.ProviderDatadog,
			"splunk",
		},
		appTheme,
	)
}

func TestModel_RoutesFromProviderToRegion(t *testing.T) {
	model := newModel()
	if !model.ApplyState(onboarding.State{NextStep: onboarding.StepDatadogRegion}) {
		t.Fatal("expected datadog region state to be handled")
	}
	if !strings.Contains(model.View().Content, "Select your integration provider:") {
		t.Fatalf("expected provider select view, got %q", model.View().Content)
	}

	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected provider select command")
	}
	_, _ = model.Update(cmd())

	if !strings.Contains(model.View().Content, "Select your Datadog site:") {
		t.Fatalf("expected datadog region view after provider select, got %q", model.View().Content)
	}
}

func TestModel_ProviderSelectionActivatesDatadogRegionInput(t *testing.T) {
	model := newModel()
	if !model.ApplyState(onboarding.State{NextStep: onboarding.StepDatadogRegion}) {
		t.Fatal("expected datadog region state to be handled")
	}

	// First enter confirms provider.
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected provider select command")
	}
	_, _ = model.Update(cmd())

	// Next enter must come from Datadog region screen.
	_, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected datadog region command")
	}
	msg := cmd()
	if _, ok := msg.(SetDatadogSiteMsg); !ok {
		t.Fatalf("expected SetDatadogSiteMsg, got %T", msg)
	}
}

func TestModel_DiscoveryRefreshCommand(t *testing.T) {
	model := newModel()
	if !model.ApplyState(onboarding.State{NextStep: onboarding.StepDatadogDiscovery}) {
		t.Fatal("expected discovery state to be handled")
	}
	_, cmd := model.Update(tea.KeyPressMsg{Text: "r", Code: 'r'})
	if cmd == nil {
		t.Fatal("expected refresh command")
	}
	msg := cmd()
	if _, ok := msg.(RefreshRequestedMsg); !ok {
		t.Fatalf("expected RefreshRequestedMsg, got %T", msg)
	}
}
