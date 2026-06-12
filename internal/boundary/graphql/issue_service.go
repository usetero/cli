package graphql

import (
	"context"

	"github.com/usetero/cli/internal/boundary/graphql/gen"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
)

// Issues provides access to the account's active issues.
type Issues interface {
	GetSummary(ctx context.Context) (domain.IssueSummary, error)
}

// IssueService reads issue aggregates from the control plane.
type IssueService struct {
	client Client
	scope  log.Scope
}

var _ Issues = (*IssueService)(nil)

// NewIssueService creates a new issue service.
func NewIssueService(client Client, scope log.Scope) *IssueService {
	return &IssueService{
		client: client,
		scope:  scope.Child("issues"),
	}
}

// GetSummary fetches the server-computed active-issue summary, including the
// per-priority breakdown. Counts are never aggregated locally.
func (s *IssueService) GetSummary(ctx context.Context) (domain.IssueSummary, error) {
	s.scope.Debug("fetching issue summary from API")
	resp, err := s.client.GetIssueSummary(ctx)
	if err != nil {
		s.scope.Error("failed to fetch issue summary", "error", err)
		return domain.IssueSummary{}, err
	}

	summary := domain.IssueSummary{
		Open: int64(resp.Issues.Summary.Count),
	}
	for _, bucket := range resp.Issues.Facets.Priorities.Buckets {
		switch bucket.Value {
		case gen.IssuePriorityHigh:
			summary.HighCount = int64(bucket.Count)
		case gen.IssuePriorityMedium:
			summary.MediumCount = int64(bucket.Count)
		case gen.IssuePriorityLow:
			summary.LowCount = int64(bucket.Count)
		}
	}

	s.scope.Debug("fetched issue summary",
		log.Int("open", int(summary.Open)),
		log.Int("high", int(summary.HighCount)))
	return summary, nil
}
