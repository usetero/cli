package integrations

import controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"

var (
	_ DatadogService      = (*RemoteDatadogService)(nil)
	_ remoteDatadogClient = (*controlplane.Client)(nil)
)
