package graphql

import (
	"context"

	"github.com/usetero/cli/internal/boundary/graphql/gen"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
)

// maxServiceStatuses bounds the services-list read for the data-plane surface.
const maxServiceStatuses = 50

// Status provides the account- and service-level status reads that back the
// data-plane status surfaces. All aggregates are computed by the control plane.
type Status interface {
	GetAccountSummary(ctx context.Context) (domain.AccountSummary, error)
	ListServiceStatuses(ctx context.Context) ([]domain.ServiceStatus, error)
	ListServiceLogEvents(ctx context.Context, serviceID string) ([]domain.LogEventStatus, error)
}

// StatusService reads control-plane status projections.
type StatusService struct {
	client Client
	scope  log.Scope
}

var _ Status = (*StatusService)(nil)

// NewStatusService creates a new status service.
func NewStatusService(client Client, scope log.Scope) *StatusService {
	return &StatusService{
		client: client,
		scope:  scope.Child("status"),
	}
}

// GetAccountSummary fetches the account-level status summary. Returns a zero
// summary (ServiceCount 0) when no Datadog account status exists yet.
func (s *StatusService) GetAccountSummary(ctx context.Context) (domain.AccountSummary, error) {
	resp, err := s.client.GetAccountStatusSummary(ctx)
	if err != nil {
		s.scope.Error("failed to fetch account status summary", "error", err)
		return domain.AccountSummary{}, err
	}

	if len(resp.DatadogAccounts.Edges) == 0 {
		return domain.AccountSummary{}, nil
	}
	status := resp.DatadogAccounts.Edges[0].Node.Status
	if status == nil {
		return domain.AccountSummary{}, nil
	}

	cov := status.Coverage
	cur := status.Current
	summary := domain.AccountSummary{
		ReadyForUse:      status.Readiness.ReadyForUse,
		Health:           serviceHealth(status.Health),
		ServiceCount:     int64(cov.LogServiceCount),
		ActiveServices:   int64(cov.LogActiveServices),
		OkServices:       int64(cov.OkServices),
		DisabledServices: int64(cov.DisabledServices),
		InactiveServices: int64(cov.InactiveServices),
		EventCount:       int64(cov.LogEventCount),
		AnalyzedCount:    int64(cov.LogEventAnalyzedCount),
		// Service-level ground truth.
		TotalServiceVolumePerHour: cur.Services.EventsPerHour,
		TotalServiceCostPerHour:   cur.Services.VolumeUsdPerHour,
		// Discovered log-event throughput.
		TotalVolumePerHour: cur.Totals.EventsPerHour,
		TotalBytesPerHour:  cur.Totals.BytesPerHour,
		TotalCostPerHour:   cur.Totals.TotalUsdPerHour,
	}
	return summary, nil
}

// ListServiceStatuses fetches enabled services with their list-status summary.
func (s *StatusService) ListServiceStatuses(ctx context.Context) ([]domain.ServiceStatus, error) {
	resp, err := s.client.ListServiceStatuses(ctx, maxServiceStatuses)
	if err != nil {
		s.scope.Error("failed to fetch service statuses", "error", err)
		return nil, err
	}

	statuses := make([]domain.ServiceStatus, 0, len(resp.Services.Edges))
	for _, edge := range resp.Services.Edges {
		node := edge.Node
		svc := domain.ServiceStatus{ID: node.Id, Name: node.Name, Health: domain.ServiceHealthInactive}
		if node.StatusSummary != nil {
			sum := node.StatusSummary
			cur := sum.Current
			sev := cur.Severity
			svc.Health = serviceHealth(sum.Health)
			svc.LogEventCount = int64(sum.LogEventCount)
			svc.LogEventAnalyzedCount = int64(sum.LogEventAnalyzedCount)
			svc.ServiceVolumePerHour = cur.EventsPerHour
			svc.LogEventVolumePerHour = cur.EventsPerHour
			svc.LogEventBytesPerHour = cur.BytesPerHour
			svc.ServiceCostPerHourVolumeUSD = cur.TotalUsdPerHour
			svc.LogEventCostPerHourUSD = cur.TotalUsdPerHour
			svc.ServiceDebugVolumePerHour = sev.DebugEventsPerHour
			svc.ServiceInfoVolumePerHour = sev.InfoEventsPerHour
			svc.ServiceWarnVolumePerHour = sev.WarnEventsPerHour
			svc.ServiceErrorVolumePerHour = sev.ErrorEventsPerHour
			svc.ServiceOtherVolumePerHour = sev.OtherEventsPerHour
		}
		statuses = append(statuses, svc)
	}

	s.scope.Debug("fetched service statuses", log.Int("count", len(statuses)))
	return statuses, nil
}

// ListServiceLogEvents fetches the log events for a single service.
func (s *StatusService) ListServiceLogEvents(ctx context.Context, serviceID string) ([]domain.LogEventStatus, error) {
	resp, err := s.client.ListServiceLogEvents(ctx, serviceID, 25)
	if err != nil {
		s.scope.Error("failed to fetch service log events", "error", err, "serviceID", serviceID)
		return nil, err
	}

	events := make([]domain.LogEventStatus, 0, len(resp.LogEvents.Edges))
	for _, edge := range resp.LogEvents.Edges {
		node := edge.Node
		name := node.Name
		if node.DisplayName != nil && *node.DisplayName != "" {
			name = *node.DisplayName
		}
		event := domain.LogEventStatus{Name: name}
		if node.Status != nil {
			cur := node.Status.Current
			event.VolumePerHour = cur.EventsPerHour
			event.BytesPerHour = cur.BytesPerHour
			event.CostPerHourUSD = cur.TotalUsdPerHour
		}
		events = append(events, event)
	}

	s.scope.Debug("fetched service log events", log.Int("count", len(events)), log.String("serviceID", serviceID))
	return events, nil
}

// serviceHealth maps a control-plane StatusHealth to the domain health enum.
func serviceHealth(h gen.StatusHealth) domain.ServiceHealth {
	switch h {
	case gen.StatusHealthOk:
		return domain.ServiceHealthOK
	case gen.StatusHealthError:
		return domain.ServiceHealthError
	case gen.StatusHealthDisabled:
		return domain.ServiceHealthDisabled
	case gen.StatusHealthInactive:
		return domain.ServiceHealthInactive
	default:
		return domain.ServiceHealthInactive
	}
}
