package integrations

import (
	"context"

	"github.com/usetero/cli/internal/domains/tenancy"
)

type DatadogSite string

const (
	DatadogSiteUS1    DatadogSite = "US1"
	DatadogSiteUS3    DatadogSite = "US3"
	DatadogSiteUS5    DatadogSite = "US5"
	DatadogSiteEU1    DatadogSite = "EU1"
	DatadogSiteUS1Fed DatadogSite = "US1_FED"
	DatadogSiteAP1    DatadogSite = "AP1"
	DatadogSiteAP2    DatadogSite = "AP2"
)

func (s DatadogSite) Valid() bool {
	switch s {
	case DatadogSiteUS1, DatadogSiteUS3, DatadogSiteUS5, DatadogSiteEU1, DatadogSiteUS1Fed, DatadogSiteAP1, DatadogSiteAP2:
		return true
	default:
		return false
	}
}

type DatadogAccountID string

type DatadogAccount struct {
	ID   DatadogAccountID
	Name string
	Site DatadogSite
}

type DatadogAccountHealth string

const (
	DatadogHealthDisabled DatadogAccountHealth = "DISABLED"
	DatadogHealthInactive DatadogAccountHealth = "INACTIVE"
	DatadogHealthError    DatadogAccountHealth = "ERROR"
	DatadogHealthOK       DatadogAccountHealth = "OK"
)

type DatadogStatus struct {
	Health               DatadogAccountHealth
	ReadyForUse          bool
	ServiceCount         int
	ActiveServices       int
	OKServices           int
	DisabledServices     int
	InactiveServices     int
	EventCount           int
	AnalyzedCount        int
	PendingPolicyCount   int
	ApprovedPolicyCount  int
	DismissedPolicyCount int
}

type CreateDatadogAccountInput struct {
	AccountID tenancy.AccountID
	Name      string
	Site      DatadogSite
	APIKey    string
	AppKey    string
}

// DatadogService is the domain contract for Datadog onboarding operations.
type DatadogService interface {
	GetByAccount(ctx context.Context, accountID tenancy.AccountID) (*DatadogAccount, error)
	ValidateAPIKey(ctx context.Context, site DatadogSite, apiKey string) (bool, string, error)
	Create(ctx context.Context, input CreateDatadogAccountInput) (DatadogAccountID, error)
	Status(ctx context.Context, datadogAccountID DatadogAccountID) (*DatadogStatus, error)
}
