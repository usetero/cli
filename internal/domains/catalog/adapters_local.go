package catalog

import (
	"time"

	logeventsdb "github.com/usetero/cli/internal/domains/catalog/db/logeventsgen"
	servicesdb "github.com/usetero/cli/internal/domains/catalog/db/servicesgen"
)

func ptrString(v string) *string { return &v }

func parseTime(v *string) time.Time {
	if v == nil || *v == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, *v)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func fromServicesDBListRow(row servicesdb.ListServicesRow) Service {
	return Service{
		ID:                    ServiceID(value(row.ID)),
		AccountID:             value(row.AccountID),
		Name:                  value(row.Name),
		Enabled:               isTruthy(row.Enabled),
		InitialWeeklyLogCount: row.InitialWeeklyLogCount,
		CreatedAt:             parseTime(row.CreatedAt),
	}
}

func fromServicesDBFactRow(row servicesdb.ListAllServiceFactsRow) ServiceFactsRow {
	return ServiceFactsRow{
		Namespace: value(row.Namespace),
		FactType:  value(row.FactType),
		Value:     value(row.Value),
	}
}

func fromLogEventsDBListByServiceRow(row logeventsdb.ListLogEventsByServiceRow) LogEvent {
	return LogEvent{
		ID:                    LogEventID(value(row.ID)),
		AccountID:             value(row.AccountID),
		ServiceID:             ServiceID(value(row.ServiceID)),
		Name:                  value(row.Name),
		Description:           value(row.Description),
		Severity:              value(row.Severity),
		BaselineAvgBytes:      row.BaselineAvgBytes,
		BaselineVolumePerHour: row.BaselineVolumePerHour,
		CreatedAt:             parseTime(row.CreatedAt),
	}
}

func fromLogEventsDBFactByServiceRow(row logeventsdb.ListLogEventFactsByServiceRow) LogEventFactsRow {
	return LogEventFactsRow{
		Namespace: value(row.SliceName),
		FactType:  value(row.FactName),
		Value:     value(row.Value),
	}
}

func value(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func valueInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func isTruthy(v *int64) bool {
	return valueInt64(v) != 0
}
