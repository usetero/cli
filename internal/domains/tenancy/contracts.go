package tenancy

import controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"

var (
	_ AccountService      = (*RemoteAccountService)(nil)
	_ WorkspaceService    = (*LocalWorkspaceService)(nil)
	_ WorkspaceService    = (*RemoteWorkspaceService)(nil)
	_ OrganizationService = (*RemoteOrganizationService)(nil)

	_ remoteAccountClient      = (*controlplane.Client)(nil)
	_ remoteWorkspaceClient    = (*controlplane.Client)(nil)
	_ remoteOrganizationClient = (*controlplane.Client)(nil)
)
