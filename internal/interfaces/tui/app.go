package tui

import (
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/cli/config"
	"github.com/usetero/cli/internal/interfaces/tui/components/statusbar"
	"github.com/usetero/cli/internal/interfaces/tui/filter"
	"github.com/usetero/cli/internal/interfaces/tui/root"
	onboardingscreen "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding"
	integrationsflow "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations"
	datadogapikey "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations/datadog/api_key"
	datadogappkey "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations/datadog/app_key"
	datadogregion "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations/datadog/region"
	providerselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations/provider/select"
	powersyncscreen "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/powersync"
	"github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/role"
	tenancyflow "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy"
	accountcreate "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/account/create"
	accountselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/account/select"
	organizationcreate "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/organization/create"
	organizationselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/organization/select"
	workspaceselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/workspace/select"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

// Start runs the TUI mode.
func Start(cfg config.RuntimeConfig, scope logging.Scope) error {
	appTheme := theme.Resolve(cfg.Theme.Mode)

	deps, err := newOnboardingDependencies(cfg, scope)
	if err != nil {
		return err
	}

	onboardingModel := onboardingscreen.New(
		deps.onboarding,
		deps.session,
		role.New(scope.Child("root/onboarding/role"), appTheme),
		tenancyflow.New(
			organizationselect.New(scope.Child("root/onboarding/organization/select"), appTheme),
			organizationcreate.New(scope.Child("root/onboarding/organization/create"), appTheme),
			accountselect.New(scope.Child("root/onboarding/account/select"), appTheme),
			accountcreate.New(scope.Child("root/onboarding/account/create"), appTheme),
			workspaceselect.New(scope.Child("root/onboarding/workspace/select"), appTheme),
			appTheme,
		),
		integrationsflow.New(
			providerselect.New(scope.Child("root/onboarding/integrations/provider/select"), appTheme),
			datadogregion.New(scope.Child("root/onboarding/datadog/region"), appTheme),
			datadogapikey.New(scope.Child("root/onboarding/datadog/api_key"), appTheme),
			datadogappkey.New(scope.Child("root/onboarding/datadog/app_key"), appTheme),
			[]integrations.Provider{integrations.ProviderDatadog},
			appTheme,
		),
		powersyncscreen.New(deps.session, appTheme),
		appTheme,
	)

	statusbarModel := statusbar.New(deps.session, string(cfg.Env), appTheme)
	p := tea.NewProgram(
		root.New(scope.Child("root"), onboardingModel, statusbarModel, appTheme),
		tea.WithEnvironment(os.Environ()),
		tea.WithFilter(filter.NewInputFilter()),
	)
	_, err = p.Run()
	return err
}
