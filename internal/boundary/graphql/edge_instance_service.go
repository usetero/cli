package graphql

import (
	"context"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
)

// EdgeInstances provides access to the account's edge runtime fleet.
type EdgeInstances interface {
	List(ctx context.Context) (domain.EdgeFleet, error)
}

// EdgeInstanceService reads edge instances from the control plane.
type EdgeInstanceService struct {
	client Client
	scope  log.Scope
}

var _ EdgeInstances = (*EdgeInstanceService)(nil)

// NewEdgeInstanceService creates a new edge instance service.
func NewEdgeInstanceService(client Client, scope log.Scope) *EdgeInstanceService {
	return &EdgeInstanceService{
		client: client,
		scope:  scope.Child("edge-instances"),
	}
}

// List fetches the edge instance fleet for the active account. The total is
// server-reported; recency/connectivity is derived by callers from LastSyncAt.
func (s *EdgeInstanceService) List(ctx context.Context) (domain.EdgeFleet, error) {
	s.scope.Debug("fetching edge instances from API")
	resp, err := s.client.ListEdgeInstances(ctx)
	if err != nil {
		s.scope.Error("failed to fetch edge instances", "error", err)
		return domain.EdgeFleet{}, err
	}

	fleet := domain.EdgeFleet{
		Total:     int64(resp.EdgeInstances.TotalCount),
		Instances: make([]domain.EdgeInstance, 0, len(resp.EdgeInstances.Edges)),
	}
	for _, edge := range resp.EdgeInstances.Edges {
		node := edge.Node
		namespace := ""
		if node.ServiceNamespace != nil {
			namespace = *node.ServiceNamespace
		}
		fleet.Instances = append(fleet.Instances, domain.EdgeInstance{
			ID:               node.Id,
			InstanceID:       node.InstanceID,
			ServiceName:      node.ServiceName,
			ServiceNamespace: namespace,
			LastSyncAt:       node.LastSyncAt,
		})
	}

	s.scope.Debug("fetched edge instances", log.Int("count", len(fleet.Instances)))
	return fleet, nil
}
