package cli

import (
	"github.com/usetero/cli/internal/interfaces/tui"
	"github.com/usetero/cli/internal/interfaces/tui/screens"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

func newTUIDependencies(exec *runner) (tui.Dependencies, error) {
	appTheme := theme.Resolve(exec.cfg.Theme.Mode)
	identityService, err := newIdentityService(exec)
	if err != nil {
		return tui.Dependencies{}, err
	}
	onboardingWorkflow, err := newOnboardingWorkflow(exec, identityService)
	if err != nil {
		return tui.Dependencies{}, err
	}

	body := screens.New(exec.scope.Child("tui/screens"), identityService, onboardingWorkflow, appTheme)
	return tui.Dependencies{
		Scope:                 exec.scope.Child("tui"),
		Theme:                 appTheme,
		Environment:           string(exec.cfg.Env),
		Body:                  body,
		AccountRuntimeFactory: newAccountRuntimeFactory(exec, identityService),
	}, nil
}
