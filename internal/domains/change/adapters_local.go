package change

import (
	"time"

	findingsdb "github.com/usetero/cli/internal/domains/change/db/findingsgen"
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

func fromListRow(row findingsdb.ListFindingsRow) Finding {
	return Finding{
		ID:             ID(value(row.ID)),
		AccountID:      value(row.AccountID),
		ServiceID:      ServiceID(value(row.ServiceID)),
		LogEventID:     value(row.LogEventID),
		Domain:         Domain(value(row.Domain)),
		ScopeKind:      value(row.ScopeKind),
		Type:           value(row.Type),
		ProblemVersion: valueInt64(row.ProblemVersion),
		Fingerprint:    value(row.Fingerprint),
		Details:        value(row.Details),
		ClosedAt:       parseTime(row.ClosedAt),
		CreatedAt:      parseTime(row.CreatedAt),
		Curation: Curation{
			Disposition: value(row.Disposition),
			Title:       value(row.Title),
			Body:        value(row.Body),
			Priority:    value(row.Priority),
		},
		Plan: Plan{
			Status:        value(row.PlanStatus),
			Title:         value(row.PlanTitle),
			Summary:       value(row.PlanSummary),
			Rationale:     value(row.PlanRationale),
			OpenQuestions: value(row.PlanOpenQuestions),
			Steps:         value(row.PlanSteps),
			Revision:      valueInt64(row.PlanRevision),
			Version:       valueInt64(row.PlanVersion),
			UpdatedAt:     parseTime(row.PlanUpdatedAt),
		},
		Status: Status{
			IsolatedEventsPerHour:   row.IsolatedEventsPerHour,
			IsolatedTotalUSDPerHour: row.IsolatedTotalUsdPerHour,
			LogEventCount:           row.LogEventCount,
			FindingUpdatedAt:        parseTime(row.FindingUpdatedAt),
		},
	}
}

func fromDomainRow(row findingsdb.ListFindingsByDomainRow) Finding {
	return Finding{
		ID:             ID(value(row.ID)),
		AccountID:      value(row.AccountID),
		ServiceID:      ServiceID(value(row.ServiceID)),
		LogEventID:     value(row.LogEventID),
		Domain:         Domain(value(row.Domain)),
		ScopeKind:      value(row.ScopeKind),
		Type:           value(row.Type),
		ProblemVersion: valueInt64(row.ProblemVersion),
		Fingerprint:    value(row.Fingerprint),
		Details:        value(row.Details),
		ClosedAt:       parseTime(row.ClosedAt),
		CreatedAt:      parseTime(row.CreatedAt),
		Curation: Curation{
			Disposition: value(row.Disposition),
			Title:       value(row.Title),
			Body:        value(row.Body),
			Priority:    value(row.Priority),
		},
		Plan: Plan{
			Status:        value(row.PlanStatus),
			Title:         value(row.PlanTitle),
			Summary:       value(row.PlanSummary),
			Rationale:     value(row.PlanRationale),
			OpenQuestions: value(row.PlanOpenQuestions),
			Steps:         value(row.PlanSteps),
			Revision:      valueInt64(row.PlanRevision),
			Version:       valueInt64(row.PlanVersion),
			UpdatedAt:     parseTime(row.PlanUpdatedAt),
		},
		Status: Status{
			IsolatedEventsPerHour:   row.IsolatedEventsPerHour,
			IsolatedTotalUSDPerHour: row.IsolatedTotalUsdPerHour,
			LogEventCount:           row.LogEventCount,
			FindingUpdatedAt:        parseTime(row.FindingUpdatedAt),
		},
	}
}

func fromServiceRow(row findingsdb.ListFindingsByServiceRow) Finding {
	return Finding{
		ID:             ID(value(row.ID)),
		AccountID:      value(row.AccountID),
		ServiceID:      ServiceID(value(row.ServiceID)),
		LogEventID:     value(row.LogEventID),
		Domain:         Domain(value(row.Domain)),
		ScopeKind:      value(row.ScopeKind),
		Type:           value(row.Type),
		ProblemVersion: valueInt64(row.ProblemVersion),
		Fingerprint:    value(row.Fingerprint),
		Details:        value(row.Details),
		ClosedAt:       parseTime(row.ClosedAt),
		CreatedAt:      parseTime(row.CreatedAt),
		Curation: Curation{
			Disposition: value(row.Disposition),
			Title:       value(row.Title),
			Body:        value(row.Body),
			Priority:    value(row.Priority),
		},
		Plan: Plan{
			Status:        value(row.PlanStatus),
			Title:         value(row.PlanTitle),
			Summary:       value(row.PlanSummary),
			Rationale:     value(row.PlanRationale),
			OpenQuestions: value(row.PlanOpenQuestions),
			Steps:         value(row.PlanSteps),
			Revision:      valueInt64(row.PlanRevision),
			Version:       valueInt64(row.PlanVersion),
			UpdatedAt:     parseTime(row.PlanUpdatedAt),
		},
		Status: Status{
			IsolatedEventsPerHour:   row.IsolatedEventsPerHour,
			IsolatedTotalUSDPerHour: row.IsolatedTotalUsdPerHour,
			LogEventCount:           row.LogEventCount,
			FindingUpdatedAt:        parseTime(row.FindingUpdatedAt),
		},
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
