package change

import (
	"context"
	"time"
)

type ID string
type Domain string
type ServiceID string

type Finding struct {
	ID             ID
	AccountID      string
	ServiceID      ServiceID
	LogEventID     string
	Domain         Domain
	ScopeKind      string
	Type           string
	ProblemVersion int64
	Fingerprint    string
	Details        string
	ClosedAt       time.Time
	CreatedAt      time.Time
	Curation       Curation
	Plan           Plan
	Status         Status
}

type FindingService interface {
	List(ctx context.Context) ([]Finding, error)
	ListByDomain(ctx context.Context, domain Domain) ([]Finding, error)
	ListByService(ctx context.Context, serviceID ServiceID) ([]Finding, error)
}
