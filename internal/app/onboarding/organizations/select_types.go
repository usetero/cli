package organizations

import "github.com/usetero/cli/internal/domain"

// tokenRefreshedMsg is sent when token refresh completes.
type tokenRefreshedMsg struct {
	err error
}

func findOrgByID(orgs []domain.Organization, id domain.OrganizationID) *domain.Organization {
	if id == "" {
		return nil
	}
	for _, org := range orgs {
		if org.ID == id {
			resolved := org
			return &resolved
		}
	}
	return nil
}
