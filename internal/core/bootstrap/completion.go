package bootstrap

import (
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
)

// Completion is the required onboarding payload for entering chat. The account
// is the working context; there is no workspace.
type Completion struct {
	User    *auth.User
	Org     domain.Organization
	Account domain.Account
}

// CompleteOnboarding validates bootstrap state and returns completion payload.
func CompleteOnboarding(state State) (Completion, bool) {
	if state.User == nil || state.Org == nil || state.Account == nil {
		return Completion{}, false
	}

	return Completion{
		User:    state.User,
		Org:     *state.Org,
		Account: *state.Account,
	}, true
}
