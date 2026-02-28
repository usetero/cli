package onboarding

import (
	iauth "github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
)

type onboardingState struct {
	user      *iauth.User
	org       *domain.Organization
	account   *domain.Account
	workspace *domain.Workspace
	ddSite    domain.DatadogSite
	ddAPIKey  string
	ddAccount domain.DatadogAccountID
}
