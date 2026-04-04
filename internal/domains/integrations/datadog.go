package integrations

import (
	"context"
	"time"

	"github.com/usetero/cli/internal/domains/validation"
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

// DatadogAPIKeySubmission is the onboarding API key submission input.
type DatadogAPIKeySubmission struct {
	APIKey DatadogAPIKey
}

// Validate normalizes and validates the Datadog API key submission input.
func (s DatadogAPIKeySubmission) Validate() (DatadogAPIKeySubmission, error) {
	parsedAPIKey, err := ParseDatadogAPIKey(s.APIKey.String())
	if err != nil {
		return DatadogAPIKeySubmission{}, err
	}
	s.APIKey = parsedAPIKey
	return s, nil
}

// DatadogAppKeySubmission is the onboarding Datadog account submission input.
type DatadogAppKeySubmission struct {
	Name   DatadogAccountName
	AppKey DatadogAppKey
}

// Validate normalizes and validates the Datadog app key submission input.
func (s DatadogAppKeySubmission) Validate() (DatadogAppKeySubmission, error) {
	parsedName, err := ParseDatadogAccountName(s.Name.String())
	if err != nil {
		return DatadogAppKeySubmission{}, err
	}
	parsedAppKey, err := ParseDatadogAppKey(s.AppKey.String())
	if err != nil {
		return DatadogAppKeySubmission{}, err
	}
	s.Name = parsedName
	s.AppKey = parsedAppKey
	return s, nil
}

// DatadogAPIKeyValidation is the Datadog API key validation input.
type DatadogAPIKeyValidation struct {
	Site   DatadogSite
	APIKey DatadogAPIKey
}

// Validate normalizes and validates the Datadog API key validation input.
func (v DatadogAPIKeyValidation) Validate() (DatadogAPIKeyValidation, error) {
	if !v.Site.Valid() {
		return DatadogAPIKeyValidation{}, validation.Struct(struct {
			Site string `label:"datadog site" validate:"required"`
		}{})
	}
	parsedAPIKey, err := ParseDatadogAPIKey(v.APIKey.String())
	if err != nil {
		return DatadogAPIKeyValidation{}, err
	}
	v.APIKey = parsedAPIKey
	return v, nil
}

// DatadogAccountCreate is the Datadog account creation mutation input.
type DatadogAccountCreate struct {
	Name   DatadogAccountName
	Site   DatadogSite
	APIKey DatadogAPIKey
	AppKey DatadogAppKey
}

// Validate normalizes and validates Datadog account create input.
func (c DatadogAccountCreate) Validate() (DatadogAccountCreate, error) {
	parsedName, err := ParseDatadogAccountName(c.Name.String())
	if err != nil {
		return DatadogAccountCreate{}, err
	}
	if !c.Site.Valid() {
		return DatadogAccountCreate{}, validation.Struct(struct {
			Site string `label:"datadog site" validate:"required"`
		}{})
	}
	parsedAPIKey, err := ParseDatadogAPIKey(c.APIKey.String())
	if err != nil {
		return DatadogAccountCreate{}, err
	}
	parsedAppKey, err := ParseDatadogAppKey(c.AppKey.String())
	if err != nil {
		return DatadogAccountCreate{}, err
	}
	c.Name = parsedName
	c.APIKey = parsedAPIKey
	c.AppKey = parsedAppKey
	return c, nil
}

// DatadogService is the domain contract for Datadog onboarding operations.
type DatadogService interface {
	Get(ctx context.Context) (*DatadogAccount, error)
	ValidateAPIKey(ctx context.Context, validation DatadogAPIKeyValidation) (bool, string, error)
	Create(ctx context.Context, create DatadogAccountCreate) (DatadogAccountID, error)
	Status(ctx context.Context, datadogAccountID DatadogAccountID) (*DatadogStatus, error)
}
