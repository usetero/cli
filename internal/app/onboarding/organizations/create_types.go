package organizations

import "github.com/usetero/cli/internal/api"

// orgCreatedMsg is sent when org creation completes.
type orgCreatedMsg struct {
	result *api.OrganizationBootstrapResult
	err    error
}
