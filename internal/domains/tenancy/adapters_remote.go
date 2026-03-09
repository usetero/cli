package tenancy

import controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"

func toControlPlaneOrganizationID(id OrganizationID) controlplane.OrganizationID {
	return controlplane.OrganizationID(id)
}

func toControlPlaneAccountID(id AccountID) controlplane.AccountID {
	return controlplane.AccountID(id)
}

func toControlPlaneWorkspaceID(id WorkspaceID) controlplane.WorkspaceID {
	return controlplane.WorkspaceID(id)
}

func fromControlPlaneOrganization(in controlplane.Organization) Organization {
	return Organization{
		ID:                   OrganizationID(in.ID),
		Name:                 in.Name,
		WorkosOrganizationID: in.WorkosOrganizationID,
	}
}

func fromControlPlaneAccount(in controlplane.Account) Account {
	return Account{
		ID:        AccountID(in.ID),
		Name:      in.Name,
		CreatedAt: in.CreatedAt,
	}
}

func fromControlPlaneWorkspace(in controlplane.Workspace, accountID AccountID) Workspace {
	return Workspace{
		ID:        WorkspaceID(in.ID),
		AccountID: accountID,
		Name:      in.Name,
		CreatedAt: in.CreatedAt,
	}
}

func fromControlPlaneBootstrap(in controlplane.OrganizationBootstrap) OrganizationBootstrap {
	account := fromControlPlaneAccount(in.Account)
	return OrganizationBootstrap{
		Organization: fromControlPlaneOrganization(in.Organization),
		Account:      account,
		Workspace:    fromControlPlaneWorkspace(in.Workspace, account.ID),
	}
}
