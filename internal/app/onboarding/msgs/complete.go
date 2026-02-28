package msgs

import (
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
)

// OnboardingComplete is emitted to the root model when onboarding finishes.
type OnboardingComplete struct {
	User      *auth.User
	Org       domain.Organization
	Account   domain.Account
	Workspace domain.Workspace
}
