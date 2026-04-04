package api

import (
	"context"
	"time"

	"github.com/usetero/cli/internal/infrastructure/controlplane/api/gen"
)

// TokenProvider returns a control-plane bearer token for API requests.
type TokenProvider interface {
	GetAccessToken(ctx context.Context) (string, error)
}

type OrganizationID string

func (id OrganizationID) String() string { return string(id) }

type AccountID string

func (id AccountID) String() string { return string(id) }

type WorkspaceID string

func (id WorkspaceID) String() string { return string(id) }

type DatadogAccountID string

func (id DatadogAccountID) String() string { return string(id) }

// Organization is the control-plane organization payload.
type Organization struct {
	ID                   OrganizationID
	Name                 string
	WorkosOrganizationID string
}

// Account is the control-plane account payload.
type Account struct {
	ID        AccountID
	Name      string
	CreatedAt time.Time
}

// Workspace is the control-plane workspace payload.
type Workspace struct {
	ID        WorkspaceID
	Name      string
	CreatedAt time.Time
}

// OrganizationBootstrap is the create-organization bootstrap payload.
type OrganizationBootstrap struct {
	Organization Organization
	Account      Account
	Workspace    Workspace
}

// DatadogSite is the Datadog regional site enum.
type DatadogSite string

const (
	DatadogSiteUS1    DatadogSite = DatadogSite(gen.DatadogAccountSiteUs1)
	DatadogSiteUS3    DatadogSite = DatadogSite(gen.DatadogAccountSiteUs3)
	DatadogSiteUS5    DatadogSite = DatadogSite(gen.DatadogAccountSiteUs5)
	DatadogSiteEU1    DatadogSite = DatadogSite(gen.DatadogAccountSiteEu1)
	DatadogSiteUS1Fed DatadogSite = DatadogSite(gen.DatadogAccountSiteUs1Fed)
	DatadogSiteAP1    DatadogSite = DatadogSite(gen.DatadogAccountSiteAp1)
	DatadogSiteAP2    DatadogSite = DatadogSite(gen.DatadogAccountSiteAp2)
)

func (s DatadogSite) Valid() bool {
	switch s {
	case DatadogSiteUS1, DatadogSiteUS3, DatadogSiteUS5, DatadogSiteEU1, DatadogSiteUS1Fed, DatadogSiteAP1, DatadogSiteAP2:
		return true
	default:
		return false
	}
}

// DatadogAccount is the Datadog integration payload.
type DatadogAccount struct {
	ID   DatadogAccountID
	Name string
	Site DatadogSite
}

type DatadogAccountHealth string

const (
	DatadogAccountHealthDisabled DatadogAccountHealth = DatadogAccountHealth(gen.StatusHealthDisabled)
	DatadogAccountHealthInactive DatadogAccountHealth = DatadogAccountHealth(gen.StatusHealthInactive)
	DatadogAccountHealthError    DatadogAccountHealth = DatadogAccountHealth(gen.StatusHealthError)
	DatadogAccountHealthOK       DatadogAccountHealth = DatadogAccountHealth(gen.StatusHealthOk)
)

type DatadogAccountStatus struct {
	Health                        DatadogAccountHealth
	ReadyForUse                   bool
	ServiceCount                  int
	ActiveServices                int
	OKServices                    int
	DisabledServices              int
	InactiveServices              int
	EventCount                    int
	AnalyzedCount                 int
	PreviewLogEventCount          int
	EffectiveLogEventCount        int
	CurrentEventsPerHour          *float64
	CurrentBytesPerHour           *float64
	CurrentTotalUSDPerHour        *float64
	PreviewSavedEventsPerHour     *float64
	PreviewSavedBytesPerHour      *float64
	PreviewSavedTotalUSDPerHour   *float64
	EffectiveSavedEventsPerHour   *float64
	EffectiveSavedBytesPerHour    *float64
	EffectiveSavedTotalUSDPerHour *float64
	RefreshedAt                   time.Time
}

type CreateDatadogAccountInput struct {
	Name   string
	Site   DatadogSite
	APIKey string
	AppKey string
}
