package organizations

import "github.com/usetero/cli/internal/domain"

// orgTokenRefreshedMsg is sent when token refresh completes.
type orgTokenRefreshedMsg struct {
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
