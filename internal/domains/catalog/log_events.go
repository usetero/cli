package catalog

import (
	"context"
	"time"
)

type LogEventID string

type LogEvent struct {
	ID                    LogEventID
	AccountID             string
	ServiceID             ServiceID
	Name                  string
	Description           string
	Severity              string
	BaselineAvgBytes      *float64
	BaselineVolumePerHour *float64
	CreatedAt             time.Time
	Facts                 LogEventFacts
	Matchers              LogMatchers
}

type LogEventService interface {
	ListLogEventsByService(ctx context.Context, serviceID ServiceID) ([]LogEvent, error)
}
