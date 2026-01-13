package api

import (
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
