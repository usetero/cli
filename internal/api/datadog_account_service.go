package api

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/pkg/client"
)

// DatadogAccountService handles Datadog account operations via the control plane API.
type DatadogAccountService struct {
	client Client
	logger log.Logger
}

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

// LogEventDiscoveryProgress tracks progress of log event discovery for a Datadog account.
type LogEventDiscoveryProgress struct {
	Status                 DiscoveryStatus
	WeeklyVolume           int64
	DiscoveredWeeklyVolume float64
	PercentComplete        *float64 // nullable
	LastError              string
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

// GetLogDiscoveryProgress gets the log event discovery progress for a Datadog account
func (s *DatadogAccountService) GetLogDiscoveryProgress(ctx context.Context, datadogAccountID string) (*LogEventDiscoveryProgress, error) {
	s.logger.Debug("fetching log discovery progress", "datadogAccountID", datadogAccountID)

	resp, err := s.client.GetDatadogAccountLogDiscoveryProgress(ctx, datadogAccountID)
	if err != nil {
		s.logger.Error("failed to fetch log discovery progress", "error", err)
		return nil, err
	}

	if len(resp.DatadogAccounts.Edges) == 0 {
		s.logger.Debug("no datadog account found")
		return nil, nil
	}

	progress := resp.DatadogAccounts.Edges[0].Node.LogEventDiscoveryProgress
	if progress.Status == "" {
		s.logger.Debug("no discovery progress available")
		return nil, nil
	}

	// Map percentComplete - handle nullable field
	var percentComplete *float64
	if progress.PercentComplete != 0 {
		percentComplete = &progress.PercentComplete
	}

	result := &LogEventDiscoveryProgress{
		Status:                 DiscoveryStatus(progress.Status),
		PercentComplete:        percentComplete,
		WeeklyVolume:           int64(progress.WeeklyVolume),
		DiscoveredWeeklyVolume: progress.WeeklyDiscoveredVolume,
		LastError:              progress.LastError,
	}

	percentStr := "nil"
	if result.PercentComplete != nil {
		percentStr = fmt.Sprintf("%.2f", *result.PercentComplete)
	}

	s.logger.Debug("fetched log discovery progress",
		log.String("status", string(result.Status)),
		log.String("percentComplete", percentStr))

	return result, nil
}
