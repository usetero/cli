package tenancyflow

import "github.com/usetero/cli/internal/domains/tenancy"

type OrganizationSelectedMsg struct {
	OrganizationID tenancy.OrganizationID
}

type OrganizationCreatedMsg struct {
	Create tenancy.OrganizationCreate
}

type AccountSelectedMsg struct {
	AccountID tenancy.AccountID
}

type AccountCreatedMsg struct {
	Create tenancy.AccountCreate
}

type WorkspaceSelectedMsg struct {
	WorkspaceID tenancy.WorkspaceID
}
