package organizations

import api "github.com/usetero/cli/internal/boundary/graphql"

// orgCreatedMsg is sent when org creation completes.
type orgCreatedMsg struct {
	result *api.OrganizationBootstrapResult
	err    error
}
