package api

import (
	"context"

	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/pkg/client"
)

// DatadogAccounts provides access to Datadog account operations.
type DatadogAccounts interface {
	HasAccount(ctx context.Context, accountID string) (bool, error)
	GetAccount(ctx context.Context, accountID string) (*DatadogAccount, error)
	ValidateAPIKey(ctx context.Context, apiKey, site string) (bool, string, error)
	CreateAccount(ctx context.Context, accountID, name, site, apiKey, appKey string) (*DatadogAccount, error)
	GetStatus(ctx context.Context, datadogAccountID string) (*DatadogAccountStatus, error)
}

// DatadogAccountService handles Datadog account operations via the control plane API.
type DatadogAccountService struct {
	client Client
	logger log.Logger
}

// Ensure DatadogAccountService implements DatadogAccounts.
var _ DatadogAccounts = (*DatadogAccountService)(nil)

// NewDatadogAccountService creates a new Datadog account service.
func NewDatadogAccountService(client Client, logger log.Logger) *DatadogAccountService {
	return &DatadogAccountService{
		client: client,
		logger: logger,
	}
}

// DatadogAccount is the domain model for a Datadog account.
type DatadogAccount struct {
	ID   string
	Name string
	Site string // GraphQL enum value (US1, US5, EU1, etc.)
}

// DatadogAccountStatusState represents the log discovery pipeline state.
type DatadogAccountStatusState string

const (
	DatadogAccountStatusDisabled    DatadogAccountStatusState = "DISABLED"
	DatadogAccountStatusInactive    DatadogAccountStatusState = "INACTIVE"
	DatadogAccountStatusBroken      DatadogAccountStatusState = "BROKEN"
	DatadogAccountStatusStale       DatadogAccountStatusState = "STALE"
	DatadogAccountStatusDiscovering DatadogAccountStatusState = "DISCOVERING"
	DatadogAccountStatusAnalyzing   DatadogAccountStatusState = "ANALYZING"
	DatadogAccountStatusReady       DatadogAccountStatusState = "READY"
)

// DatadogAccountStatus tracks the log discovery status for a Datadog account.
type DatadogAccountStatus struct {
	Status              DatadogAccountStatusState
	PercentComplete     float64
	ServiceLogVolume    int // Total log volume across all active services
	DiscoveredLogVolume int // Volume of logs we've analyzed
	ServiceCount        int // Total number of services
	ActiveServices      int // Services not DISABLED or INACTIVE
	ReadyServices       int
	AnalyzingServices   int
	DiscoveringServices int
	StaleServices       int
	BrokenServices      int
	DisabledServices    int
	InactiveServices    int
	SavedCount          int  // Number of log events saved (for progress display)
	ReadyForUse         bool // Whether the account has enough data to proceed
}

// HasAccount checks if an account has a Datadog integration configured
func (s *DatadogAccountService) HasAccount(ctx context.Context, accountID string) (bool, error) {
	s.logger.Debug("checking for datadog account via API", "accountID", accountID)
	resp, err := s.client.GetAccount(ctx, accountID)
	if err != nil {
		s.logger.Error("failed to check datadog account", "error", err, "accountID", accountID)
		return false, err
	}

	// Check if we found an account and if it has a datadogAccount
	// When null in GraphQL, genqlient returns empty struct with empty Id
	if len(resp.Accounts.Edges) > 0 {
		account := resp.Accounts.Edges[0].Node
		hasDatadog := account.DatadogAccount.Id != ""
		s.logger.Debug("checked for datadog account via API", "hasDatadog", hasDatadog)
		return hasDatadog, nil
	}

	s.logger.Debug("checked for datadog account via API", "hasDatadog", false)
	return false, nil
}

// GetAccount retrieves the Datadog account for the given account ID, or nil if none exists
func (s *DatadogAccountService) GetAccount(ctx context.Context, accountID string) (*DatadogAccount, error) {
	s.logger.Debug("fetching datadog account via API", "accountID", accountID)
	resp, err := s.client.GetAccount(ctx, accountID)
	if err != nil {
		s.logger.Error("failed to fetch datadog account", "error", err, "accountID", accountID)
		return nil, err
	}

	// Check if we found an account and if it has a datadogAccount
	if len(resp.Accounts.Edges) > 0 {
		account := resp.Accounts.Edges[0].Node
		if account.DatadogAccount.Id != "" {
			ddAccount := &DatadogAccount{
				ID:   account.DatadogAccount.Id,
				Name: account.DatadogAccount.Name,
				Site: string(account.DatadogAccount.Site),
			}
			s.logger.Debug("fetched datadog account via API", "datadogAccountID", ddAccount.ID)
			return ddAccount, nil
		}
	}

	s.logger.Debug("no datadog account found via API")
	return nil, nil
}

// ValidateAPIKey validates the API key via the control plane.
// The control plane handles validation against Datadog's API.
// Returns whether the key is valid, an error message if invalid, and any system errors.
func (s *DatadogAccountService) ValidateAPIKey(ctx context.Context, apiKey, site string) (bool, string, error) {
	s.logger.Debug("validating datadog API key via control plane", "site", site)

	input := client.ValidateDatadogApiKeyInput{
		ApiKey: apiKey,
		Site:   client.DatadogAccountSite(site),
	}

	resp, err := s.client.ValidateDatadogApiKey(ctx, input)
	if err != nil {
		s.logger.Error("failed to validate datadog API key", "error", err)
		return false, "", err
	}

	if !resp.ValidateDatadogApiKey.Valid {
		errorMsg := "Invalid API key"
		if resp.ValidateDatadogApiKey.Error != "" {
			errorMsg = resp.ValidateDatadogApiKey.Error
		}
		s.logger.Debug("datadog API key is invalid", "error", errorMsg)
		return false, errorMsg, nil
	}

	s.logger.Debug("validated datadog API key successfully")
	return true, "", nil
}

// CreateAccount creates a Datadog account in the control plane with credentials.
// Both API key and Application key must be provided.
// Keys are sent to control plane and stored securely there - never stored locally.
// The control plane validates the credentials before creating the account.
func (s *DatadogAccountService) CreateAccount(ctx context.Context, accountID, name, site, apiKey, appKey string) (*DatadogAccount, error) {
	s.logger.Debug("creating datadog account with credentials via API", "accountID", accountID, "site", site)
	input := client.CreateDatadogAccountWithCredentialsInput{
		Attributes: client.CreateDatadogAccountInput{
			AccountID: accountID,
			Name:      name,
			Site:      client.DatadogAccountSite(site), // US1, US5, EU1, etc.
		},
		Credentials: client.CreateDatadogCredentialsInput{
			ApiKey: apiKey,
			AppKey: appKey,
		},
	}

	resp, err := s.client.CreateDatadogAccountWithCredentials(ctx, input)
	if err != nil {
		s.logger.Error("failed to create datadog account", "error", err)
		return nil, err
	}

	s.logger.Debug("created datadog account via API", "id", resp.CreateDatadogAccount.Id, "site", string(resp.CreateDatadogAccount.Site))
	return &DatadogAccount{
		ID:   resp.CreateDatadogAccount.Id,
		Name: resp.CreateDatadogAccount.Name,
		Site: string(resp.CreateDatadogAccount.Site),
	}, nil
}

// GetStatus gets the discovery status for a Datadog account.
// This is used during onboarding to track overall progress.
func (s *DatadogAccountService) GetStatus(ctx context.Context, datadogAccountID string) (*DatadogAccountStatus, error) {
	s.logger.Debug("fetching datadog account status", "datadogAccountID", datadogAccountID)

	resp, err := s.client.GetDatadogAccountStatus(ctx, datadogAccountID)
	if err != nil {
		s.logger.Error("failed to fetch datadog account status", "error", err)
		return nil, err
	}

	if len(resp.DatadogAccounts.Edges) == 0 {
		s.logger.Debug("no datadog account found")
		return nil, nil
	}

	statusNode := resp.DatadogAccounts.Edges[0].Node.Status

	result := &DatadogAccountStatus{
		Status:              DatadogAccountStatusState(statusNode.LogStatus),
		PercentComplete:     statusNode.LogPercentComplete,
		ServiceLogVolume:    statusNode.LogServiceVolumeInWindow,
		DiscoveredLogVolume: statusNode.LogDiscoveredVolumeInWindow,
		ServiceCount:        statusNode.LogServiceCount,
		ActiveServices:      statusNode.LogActiveServices,
		ReadyServices:       statusNode.LogReadyServices,
		AnalyzingServices:   statusNode.LogAnalyzingServices,
		DiscoveringServices: statusNode.LogDiscoveringServices,
		StaleServices:       statusNode.LogStaleServices,
		BrokenServices:      statusNode.LogBrokenServices,
		DisabledServices:    statusNode.LogDisabledServices,
		InactiveServices:    statusNode.LogInactiveServices,
		SavedCount:          statusNode.LogSavedCount,
		ReadyForUse:         statusNode.ReadyForUse,
	}

	s.logger.Debug("fetched datadog account status",
		log.String("status", string(result.Status)),
		log.Int("serviceCount", result.ServiceCount),
		log.Int("ready", result.ReadyServices),
		log.Int("inactive", result.InactiveServices))

	return result, nil
}
