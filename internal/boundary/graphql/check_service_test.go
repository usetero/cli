package graphql_test

import (
	"context"
	"errors"
	"testing"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/boundary/graphql/apitest"
	"github.com/usetero/cli/internal/boundary/graphql/gen"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
)

func TestCheckService_List(t *testing.T) {
	t.Parallel()
	t.Run("maps checks, posture, and domain facets", func(t *testing.T) {
		t.Parallel()
		cost := 12.5
		mockClient := &apitest.MockClient{
			ListChecksFunc: func(ctx context.Context) (*gen.ListChecksResponse, error) {
				return &gen.ListChecksResponse{
					Checks: gen.ListChecksChecksCheckConnection{
						TotalCount: 2,
						Edges: []gen.ListChecksChecksCheckConnectionEdgesCheckEdge{
							{Node: gen.ListChecksChecksCheckConnectionEdgesCheckEdgeNodeCheck{
								Id:     "chk-1",
								Name:   "Debug noise",
								Domain: gen.FindingCheckDomainCost,
								Posture: gen.ListChecksChecksCheckConnectionEdgesCheckEdgeNodeCheckPosture{
									OpenFindingCount:     5,
									PendingFindingCount:  3,
									ActiveIssueCount:     2,
									AffectedServiceCount: 4,
									Current: gen.ListChecksChecksCheckConnectionEdgesCheckEdgeNodeCheckPostureCurrentStatusMeasurementTotals{
										TotalUsdPerHour: &cost,
									},
								},
							}},
							{Node: gen.ListChecksChecksCheckConnectionEdgesCheckEdgeNodeCheck{
								Id:     "chk-2",
								Name:   "PII exposure",
								Domain: gen.FindingCheckDomainCompliance,
							}},
						},
						Facets: gen.ListChecksChecksCheckConnectionFacetsCheckFacets{
							Domains: gen.ListChecksChecksCheckConnectionFacetsCheckFacetsDomainsCheckDomainFacet{
								Buckets: []gen.ListChecksChecksCheckConnectionFacetsCheckFacetsDomainsCheckDomainFacetBucketsCheckDomainFacetBucket{
									{Value: gen.FindingCheckDomainCost, Count: 1},
									{Value: gen.FindingCheckDomainCompliance, Count: 1},
								},
							},
						},
					},
				}, nil
			},
		}

		svc := graphql.NewCheckService(mockClient, logtest.NewScope(t))
		catalog, err := svc.List(context.Background())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if catalog.Total != 2 || len(catalog.Checks) != 2 {
			t.Fatalf("catalog = %+v, want total=2 and 2 checks", catalog)
		}
		first := catalog.Checks[0]
		if first.ID != "chk-1" || first.Domain != domain.CheckDomainCost || first.OpenFindingCount != 5 {
			t.Errorf("first check = %+v", first)
		}
		if first.CurrentCostPerHour == nil || *first.CurrentCostPerHour != 12.5 {
			t.Errorf("first check cost = %v, want 12.5", first.CurrentCostPerHour)
		}
		if catalog.DomainCount(domain.CheckDomainCost) != 1 || catalog.DomainCount(domain.CheckDomainCompliance) != 1 {
			t.Errorf("domain counts = %+v", catalog.ByDomain)
		}
	})

	t.Run("propagates client errors", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			ListChecksFunc: func(ctx context.Context) (*gen.ListChecksResponse, error) {
				return nil, errors.New("network error")
			},
		}

		svc := graphql.NewCheckService(mockClient, logtest.NewScope(t))
		_, err := svc.List(context.Background())

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
