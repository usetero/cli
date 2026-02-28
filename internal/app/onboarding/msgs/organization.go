package msgs

import "github.com/usetero/cli/internal/domain"

// OrgSelected is emitted when user selects an organization.
type OrgSelected struct {
	Org domain.Organization
}

// NoOrgs is emitted when no organizations exist.
type NoOrgs struct{}

// OrgCreated is emitted when a new organization is created.
type OrgCreated struct {
	Org domain.Organization
}
