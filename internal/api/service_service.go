package api

import (
	"github.com/usetero/cli/internal/log"
)

// Services provides access to service operations.
type Services interface {
	// No methods yet - services are discovered from observability platforms
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
