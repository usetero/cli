package graphql_test

import (
	"context"
	"errors"
	"testing"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/boundary/graphql/apitest"
	"github.com/usetero/cli/internal/boundary/graphql/gen"
	"github.com/usetero/cli/internal/log/logtest"
)

func TestIssueService_GetSummary(t *testing.T) {
	t.Parallel()
	t.Run("maps server count and priority facet buckets", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			GetIssueSummaryFunc: func(ctx context.Context) (*gen.GetIssueSummaryResponse, error) {
				return &gen.GetIssueSummaryResponse{
					Issues: gen.GetIssueSummaryIssuesIssueConnection{
						TotalCount: 9,
						Summary:    gen.GetIssueSummaryIssuesIssueConnectionSummaryIssueSummary{Count: 9},
						Facets: gen.GetIssueSummaryIssuesIssueConnectionFacetsIssueFacets{
							Priorities: gen.GetIssueSummaryIssuesIssueConnectionFacetsIssueFacetsPrioritiesIssuePriorityFacet{
								Buckets: []gen.GetIssueSummaryIssuesIssueConnectionFacetsIssueFacetsPrioritiesIssuePriorityFacetBucketsIssuePriorityFacetBucket{
									{Value: gen.IssuePriorityHigh, Count: 4},
									{Value: gen.IssuePriorityMedium, Count: 3},
									{Value: gen.IssuePriorityLow, Count: 2},
								},
							},
						},
					},
				}, nil
			},
		}

		svc := graphql.NewIssueService(mockClient, logtest.NewScope(t))
		summary, err := svc.GetSummary(context.Background())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summary.Open != 9 {
			t.Errorf("Open = %d, want 9", summary.Open)
		}
		if summary.HighCount != 4 || summary.MediumCount != 3 || summary.LowCount != 2 {
			t.Errorf("priority counts = %+v, want high=4 medium=3 low=2", summary)
		}
	})

	t.Run("propagates client errors", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			GetIssueSummaryFunc: func(ctx context.Context) (*gen.GetIssueSummaryResponse, error) {
				return nil, errors.New("network error")
			},
		}

		svc := graphql.NewIssueService(mockClient, logtest.NewScope(t))
		_, err := svc.GetSummary(context.Background())

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
