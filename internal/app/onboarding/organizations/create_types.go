package organizations

import graphql "github.com/usetero/cli/internal/boundary/graphql"

// orgCreatedMsg is sent when org creation completes.
type orgCreatedMsg struct {
	result *graphql.OrganizationBootstrapResult
	err    error
}
