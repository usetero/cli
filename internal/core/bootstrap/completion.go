package bootstrap

import (
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
)

// Completion is the required onboarding payload for entering chat.
type Completion struct {
	User      *auth.User
	Org       domain.Organization
	Account   domain.Account
	Workspace domain.Workspace
}

// CompleteOnboarding validates bootstrap state and returns completion payload.
func CompleteOnboarding(state State) (Completion, bool) {
	if state.User == nil || state.Org == nil || state.Account == nil || state.Workspace == nil {
		return Completion{}, false
	}

	return Completion{
		User:      state.User,
		Org:       *state.Org,
		Account:   *state.Account,
		Workspace: *state.Workspace,
	}, true
}
