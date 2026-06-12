package graphql

import (
	"context"

	"github.com/usetero/cli/internal/boundary/graphql/gen"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
)

// Checks provides access to the code-defined product check catalog.
type Checks interface {
	List(ctx context.Context) (domain.CheckCatalog, error)
}

// CheckService reads product checks and their account-scoped posture from the
// control plane.
type CheckService struct {
	client Client
	scope  log.Scope
}

var _ Checks = (*CheckService)(nil)

// NewCheckService creates a new check service.
func NewCheckService(client Client, scope log.Scope) *CheckService {
	return &CheckService{
		client: client,
		scope:  scope.Child("checks"),
	}
}

// List fetches all product checks with their posture and the server-computed
// per-domain counts.
func (s *CheckService) List(ctx context.Context) (domain.CheckCatalog, error) {
	s.scope.Debug("fetching checks from API")
	resp, err := s.client.ListChecks(ctx)
	if err != nil {
		s.scope.Error("failed to fetch checks", "error", err)
		return domain.CheckCatalog{}, err
	}

	catalog := domain.CheckCatalog{
		Total:    int64(resp.Checks.TotalCount),
		Checks:   make([]domain.Check, 0, len(resp.Checks.Edges)),
		ByDomain: make(map[domain.CheckDomain]int64),
	}
	for _, edge := range resp.Checks.Edges {
		node := edge.Node
		catalog.Checks = append(catalog.Checks, domain.Check{
			ID:                    node.Id,
			Name:                  node.Name,
			Domain:                checkDomain(node.Domain),
			OpenFindingCount:      int64(node.Posture.OpenFindingCount),
			PendingFindingCount:   int64(node.Posture.PendingFindingCount),
			EscalatedFindingCount: int64(node.Posture.EscalatedFindingCount),
			ActiveIssueCount:      int64(node.Posture.ActiveIssueCount),
			AffectedServiceCount:  int64(node.Posture.AffectedServiceCount),
			CurrentCostPerHour:    node.Posture.Current.TotalUsdPerHour,
		})
	}
	for _, bucket := range resp.Checks.Facets.Domains.Buckets {
		catalog.ByDomain[checkDomain(bucket.Value)] = int64(bucket.Count)
	}

	s.scope.Debug("fetched checks", log.Int("count", len(catalog.Checks)))
	return catalog, nil
}

func checkDomain(d gen.FindingCheckDomain) domain.CheckDomain {
	switch d {
	case gen.FindingCheckDomainCost:
		return domain.CheckDomainCost
	case gen.FindingCheckDomainCompliance:
		return domain.CheckDomainCompliance
	default:
		return domain.CheckDomain(d)
	}
}
