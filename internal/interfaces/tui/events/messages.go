package events

import (
	"github.com/usetero/cli/internal/domains/tenancy"
	accountruntime "github.com/usetero/cli/internal/runtime/account"
)

// OrganizationSelectedMsg is emitted when onboarding or navigation selects an organization.
type OrganizationSelectedMsg struct {
	Organization tenancy.Organization
}

// AccountSelectedMsg is emitted when onboarding or navigation selects an account scope.
type AccountSelectedMsg struct {
	Scope accountruntime.Scope
}

// AccountRuntimeUpdatedMsg carries the latest account runtime status through the app message bus.
type AccountRuntimeUpdatedMsg struct {
	Status accountruntime.Status
}

// OnboardingCompletedMsg notifies the app body router that bootstrap completed
// and the main product surface should become active.
type OnboardingCompletedMsg struct{}
