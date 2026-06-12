package graphql_test

import (
	"context"
	"errors"
	"testing"
	"time"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/boundary/graphql/apitest"
	"github.com/usetero/cli/internal/boundary/graphql/gen"
	"github.com/usetero/cli/internal/log/logtest"
)

func TestEdgeInstanceService_List(t *testing.T) {
	t.Parallel()
	t.Run("maps fleet total and instances", func(t *testing.T) {
		t.Parallel()
		ns := "payments"
		now := time.Now()
		mockClient := &apitest.MockClient{
			ListEdgeInstancesFunc: func(ctx context.Context) (*gen.ListEdgeInstancesResponse, error) {
				return &gen.ListEdgeInstancesResponse{
					EdgeInstances: gen.ListEdgeInstancesEdgeInstancesEdgeInstanceConnection{
						TotalCount: 2,
						Edges: []gen.ListEdgeInstancesEdgeInstancesEdgeInstanceConnectionEdgesEdgeInstanceEdge{
							{Node: gen.ListEdgeInstancesEdgeInstancesEdgeInstanceConnectionEdgesEdgeInstanceEdgeNodeEdgeInstance{
								Id:               "edge-1",
								InstanceID:       "inst-1",
								ServiceName:      "checkout",
								ServiceNamespace: &ns,
								LastSyncAt:       now,
							}},
							{Node: gen.ListEdgeInstancesEdgeInstancesEdgeInstanceConnectionEdgesEdgeInstanceEdgeNodeEdgeInstance{
								Id:          "edge-2",
								InstanceID:  "inst-2",
								ServiceName: "billing",
								LastSyncAt:  now.Add(-time.Hour),
							}},
						},
					},
				}, nil
			},
		}

		svc := graphql.NewEdgeInstanceService(mockClient, logtest.NewScope(t))
		fleet, err := svc.List(context.Background())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fleet.Total != 2 || len(fleet.Instances) != 2 {
			t.Fatalf("fleet = %+v, want total=2 and 2 instances", fleet)
		}
		if fleet.Instances[0].ServiceNamespace != "payments" {
			t.Errorf("namespace = %q, want payments", fleet.Instances[0].ServiceNamespace)
		}
		if fleet.Instances[1].ServiceNamespace != "" {
			t.Errorf("nil namespace should map to empty string, got %q", fleet.Instances[1].ServiceNamespace)
		}
		if got := fleet.ConnectedCount(now, 30*time.Minute); got != 1 {
			t.Errorf("ConnectedCount = %d, want 1", got)
		}
	})

	t.Run("propagates client errors", func(t *testing.T) {
		t.Parallel()
		mockClient := &apitest.MockClient{
			ListEdgeInstancesFunc: func(ctx context.Context) (*gen.ListEdgeInstancesResponse, error) {
				return nil, errors.New("network error")
			},
		}

		svc := graphql.NewEdgeInstanceService(mockClient, logtest.NewScope(t))
		_, err := svc.List(context.Background())

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
