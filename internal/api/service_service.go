package api

import (
	"context"
	"time"

	"github.com/usetero/cli/internal/log"
)

// ServiceService handles service-related operations.
// Services are discovered from observability platforms (Datadog, Splunk, etc.)
// and represent applications/microservices generating telemetry.
type ServiceService struct {
	client Client
	logger log.Logger
}

// NewServiceService creates a new service service.
func NewServiceService(client Client, logger log.Logger) *ServiceService {
	return &ServiceService{
		client: client,
		logger: logger,
	}
}

// ServiceDiscoveryStatus tracks the status of service discovery for a Datadog account.
type ServiceDiscoveryStatus struct {
	Status              DiscoveryStatus
	ServicesDiscovered  int
	LastError           string
	StartedAt           *time.Time
	CompletedAt         *time.Time
	ConsecutiveFailures int
}

// DiscoveryStatus represents the state of a discovery job.
type DiscoveryStatus string

const (
	DiscoveryStatusPending     DiscoveryStatus = "PENDING"
	DiscoveryStatusDiscovering DiscoveryStatus = "DISCOVERING"
	DiscoveryStatusReady       DiscoveryStatus = "READY"
	DiscoveryStatusError       DiscoveryStatus = "ERROR"
)

// GetServiceDiscoveryStatus checks if service discovery has completed for a Datadog account.
// Returns the embedded ServiceDiscoveryStatus from the DatadogAccount.
func (s *ServiceService) GetServiceDiscoveryStatus(ctx context.Context, datadogAccountID string) (*ServiceDiscoveryStatus, error) {
	s.logger.Debug("getting service discovery status", "datadogAccountID", datadogAccountID)

	resp, err := s.client.GetDatadogAccountServiceDiscoveryProgress(ctx, datadogAccountID)
	if err != nil {
		s.logger.Error("failed to get service discovery status", "error", err)
		return nil, err
	}

	// Return first result if available
	if len(resp.DatadogAccounts.Edges) > 0 {
		node := resp.DatadogAccounts.Edges[0].Node
		progressNode := node.ServiceDiscoveryProgress

		// Map startedAt and completedAt if present
		var startedAt *time.Time
		if !progressNode.StartedAt.IsZero() {
			startedAt = &progressNode.StartedAt
		}
		var completedAt *time.Time
		if !progressNode.CompletedAt.IsZero() {
			completedAt = &progressNode.CompletedAt
		}

		status := &ServiceDiscoveryStatus{
			Status:              DiscoveryStatus(progressNode.Status),
			ServicesDiscovered:  progressNode.ServicesDiscovered,
			LastError:           progressNode.LastError,
			StartedAt:           startedAt,
			CompletedAt:         completedAt,
			ConsecutiveFailures: progressNode.ConsecutiveFailures,
		}

		s.logger.Debug("got service discovery status",
			log.String("status", string(status.Status)),
			log.Int("servicesDiscovered", status.ServicesDiscovered))

		return status, nil
	}

	s.logger.Debug("no Datadog account found")
	return nil, nil
}
