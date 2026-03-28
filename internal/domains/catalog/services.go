package catalog

import (
	"context"
	"time"
)

type ServiceID string

type Service struct {
	ID                    ServiceID
	AccountID             string
	Name                  string
	Enabled               bool
	InitialWeeklyLogCount *int64
	CreatedAt             time.Time
	Facts                 ServiceFacts
}

type CatalogService interface {
	ListServices(ctx context.Context) ([]Service, error)
	ListLogEventsByService(ctx context.Context, serviceID ServiceID) ([]LogEvent, error)
}
