package api

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/log"
)

// Services provides access to service operations.
type Services interface {
	EnableService(ctx context.Context, serviceID string) error
	DisableService(ctx context.Context, serviceID string) error
}

// ServiceService handles service-related operations.
// Services are discovered from observability platforms (Datadog, Splunk, etc.)
// and represent applications/microservices generating telemetry.
type ServiceService struct {
	client Client
	scope  log.Scope
}

// Ensure ServiceService implements Services.
var _ Services = (*ServiceService)(nil)

// NewServiceService creates a new service service.
func NewServiceService(client Client, scope log.Scope) *ServiceService {
	return &ServiceService{
		client: client,
		scope:  scope.Child("services"),
	}
}

// EnableService enables a service for analysis.
func (s *ServiceService) EnableService(ctx context.Context, serviceID string) error {
	_, err := s.client.EnableService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("enable service %s: %w", serviceID, err)
	}
	return nil
}

// DisableService disables a service.
func (s *ServiceService) DisableService(ctx context.Context, serviceID string) error {
	_, err := s.client.DisableService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("disable service %s: %w", serviceID, err)
	}
	return nil
}
