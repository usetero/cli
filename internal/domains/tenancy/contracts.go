package tenancy

import controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"

var (
	_ AccountService      = (*RemoteAccountService)(nil)
	_ WorkspaceService    = (*LocalWorkspaceService)(nil)
	_ WorkspaceService    = (*RemoteWorkspaceService)(nil)
	_ OrganizationService = (*RemoteOrganizationService)(nil)

	_ remoteAccountClient      = (*controlplane.BootstrapClient)(nil)
	_ remoteWorkspaceClient    = (*controlplane.AccountClient)(nil)
	_ remoteOrganizationClient = (*controlplane.BootstrapClient)(nil)
)
