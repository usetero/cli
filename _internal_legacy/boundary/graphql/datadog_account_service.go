package graphql

import (
	"context"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/boundary/graphql/gen"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
)

// ValidateAPIKeyInput contains the fields for validating a Datadog API key.
type ValidateAPIKeyInput struct {
	APIKey string
	Site   string
}

// CreateDatadogAccountInput contains the fields for creating a Datadog account with credentials.
type CreateDatadogAccountInput struct {
	ID        uuid.UUID
	AccountID domain.AccountID
	Name      string
	Site      string
	APIKey    string
	AppKey    string
}

// DatadogAccounts provides access to Datadog account operations.
type DatadogAccounts interface {
	HasAccount(ctx context.Context, accountID domain.AccountID) (bool, error)
	GetAccount(ctx context.Context, accountID domain.AccountID) (*DatadogAccount, error)
	ValidateAPIKey(ctx context.Context, input ValidateAPIKeyInput) (bool, string, error)
	CreateAccount(ctx context.Context, input CreateDatadogAccountInput) (*DatadogAccount, error)
	GetStatus(ctx context.Context, datadogAccountID domain.DatadogAccountID) (*DatadogAccountStatus, error)
}

// DatadogAccountService handles Datadog account operations via the control plane API.
type DatadogAccountService struct {
	client Client
	scope  log.Scope
}

// Ensure DatadogAccountService implements DatadogAccounts.
var _ DatadogAccounts = (*DatadogAccountService)(nil)

// NewDatadogAccountService creates a new Datadog account service.
func NewDatadogAccountService(client Client, scope log.Scope) *DatadogAccountService {
	return &DatadogAccountService{
		client: client,
		scope:  scope.Child("datadog"),
	}
}

// DatadogAccount is the domain model for a Datadog account.
type DatadogAccount struct {
	ID   domain.DatadogAccountID
	Name string
	Site string // GraphQL enum value (US1, US5, EU1, etc.)
}

// DatadogAccountHealth represents the infrastructure health of a Datadog account.
type DatadogAccountHealth string

const (
	DatadogAccountHealthDisabled DatadogAccountHealth = "DISABLED"
	DatadogAccountHealthInactive DatadogAccountHealth = "INACTIVE"
	DatadogAccountHealthOK       DatadogAccountHealth = "OK"
)

// DatadogAccountStatus tracks the health and progress for a Datadog account.
type DatadogAccountStatus struct {
	Health               DatadogAccountHealth
	ReadyForUse          bool
	ServiceCount         int
	ActiveServices       int
	OkServices           int
	DisabledServices     int
	InactiveServices     int
	EventCount           int
	AnalyzedCount        int
	PendingPolicyCount   int
	ApprovedPolicyCount  int
	DismissedPolicyCount int
}

// HasAccount checks if an account has a Datadog integration configured
func (s *DatadogAccountService) HasAccount(ctx context.Context, accountID domain.AccountID) (bool, error) {
	s.scope.Debug("checking for datadog account via API", "accountID", accountID)
	resp, err := s.client.GetAccount(ctx, accountID.String())
	if err != nil {
		s.scope.Error("failed to check datadog account", "error", err, "accountID", accountID)
		return false, err
	}

	// Check if we found an account and if it has a datadogAccount
	if len(resp.Accounts.Edges) > 0 {
		account := resp.Accounts.Edges[0].Node
		hasDatadog := account.DatadogAccount != nil && account.DatadogAccount.Id != ""
		s.scope.Debug("checked for datadog account via API", "hasDatadog", hasDatadog)
		return hasDatadog, nil
	}

	s.scope.Debug("checked for datadog account via API", "hasDatadog", false)
	return false, nil
}

// GetAccount retrieves the Datadog account for the given account ID, or nil if none exists
func (s *DatadogAccountService) GetAccount(ctx context.Context, accountID domain.AccountID) (*DatadogAccount, error) {
	s.scope.Debug("fetching datadog account via API", "accountID", accountID)
	resp, err := s.client.GetAccount(ctx, accountID.String())
	if err != nil {
		s.scope.Error("failed to fetch datadog account", "error", err, "accountID", accountID)
		return nil, err
	}

	// Check if we found an account and if it has a datadogAccount
	if len(resp.Accounts.Edges) > 0 {
		account := resp.Accounts.Edges[0].Node
		if account.DatadogAccount != nil && account.DatadogAccount.Id != "" {
			ddAccount := &DatadogAccount{
				ID:   domain.DatadogAccountID(account.DatadogAccount.Id),
				Name: account.DatadogAccount.Name,
				Site: string(account.DatadogAccount.Site),
			}
			s.scope.Debug("fetched datadog account via API", "datadogAccountID", ddAccount.ID)
			return ddAccount, nil
		}
	}

	s.scope.Debug("no datadog account found via API")
	return nil, nil
}

// ValidateAPIKey validates the API key via the control plane.
// The control plane handles validation against Datadog's API.
// Returns whether the key is valid, an error message if invalid, and any system errors.
func (s *DatadogAccountService) ValidateAPIKey(ctx context.Context, input ValidateAPIKeyInput) (bool, string, error) {
	s.scope.Debug("validating datadog API key via control plane", "site", input.Site)

	genInput := gen.ValidateDatadogApiKeyInput{
		ApiKey: input.APIKey,
		Site:   gen.DatadogAccountSite(input.Site),
	}

	resp, err := s.client.ValidateDatadogApiKey(ctx, genInput)
	if err != nil {
		s.scope.Error("failed to validate datadog API key", "error", err)
		return false, "", err
	}

	if !resp.ValidateDatadogApiKey.Valid {
		errorMsg := "Invalid API key"
		if resp.ValidateDatadogApiKey.Error != nil && *resp.ValidateDatadogApiKey.Error != "" {
			errorMsg = *resp.ValidateDatadogApiKey.Error
		}
		s.scope.Debug("datadog API key is invalid", "error", errorMsg)
		return false, errorMsg, nil
	}

	s.scope.Debug("validated datadog API key successfully")
	return true, "", nil
}

// CreateAccount creates a Datadog account in the control plane with credentials.
// Both API key and Application key must be provided.
// Keys are sent to control plane and stored securely there - never stored locally.
// The control plane validates the credentials before creating the account.
func (s *DatadogAccountService) CreateAccount(ctx context.Context, input CreateDatadogAccountInput) (*DatadogAccount, error) {
	s.scope.Debug("creating datadog account with credentials via API", "id", input.ID.String(), "accountID", input.AccountID, "site", input.Site)
	genInput := gen.CreateDatadogAccountWithCredentialsInput{
		Attributes: gen.CreateDatadogAccountInput{
			Id:        ptr(input.ID.String()),
			AccountID: input.AccountID.String(),
			Name:      input.Name,
			Site:      gen.DatadogAccountSite(input.Site),
		},
		Credentials: gen.CreateDatadogCredentialsInput{
			ApiKey: input.APIKey,
			AppKey: input.AppKey,
		},
	}

	resp, err := s.client.CreateDatadogAccountWithCredentials(ctx, genInput)
	if err != nil {
		s.scope.Error("failed to create datadog account", "error", err)
		return nil, err
	}

	s.scope.Debug("created datadog account via API", "id", resp.CreateDatadogAccount.Id, "site", string(resp.CreateDatadogAccount.Site))
	return &DatadogAccount{
		ID:   domain.DatadogAccountID(resp.CreateDatadogAccount.Id),
		Name: resp.CreateDatadogAccount.Name,
		Site: string(resp.CreateDatadogAccount.Site),
	}, nil
}

// GetStatus gets the health status for a Datadog account.
// This is used during onboarding to track overall progress.
func (s *DatadogAccountService) GetStatus(ctx context.Context, datadogAccountID domain.DatadogAccountID) (*DatadogAccountStatus, error) {
	s.scope.Debug("fetching datadog account status", "datadogAccountID", datadogAccountID)

	resp, err := s.client.GetDatadogAccountStatus(ctx, datadogAccountID.String())
	if err != nil {
		s.scope.Error("failed to fetch datadog account status", "error", err)
		return nil, err
	}

	if len(resp.DatadogAccounts.Edges) == 0 {
		s.scope.Debug("no datadog account found")
		return nil, nil
	}

	statusNode := resp.DatadogAccounts.Edges[0].Node.Status
	if statusNode == nil {
		s.scope.Debug("status cache not populated yet")
		return nil, nil
	}

	result := &DatadogAccountStatus{
		Health:               DatadogAccountHealth(statusNode.Health),
		ReadyForUse:          statusNode.ReadyForUse,
		ServiceCount:         statusNode.LogServiceCount,
		ActiveServices:       statusNode.LogActiveServices,
		OkServices:           statusNode.OkServices,
		DisabledServices:     statusNode.DisabledServices,
		InactiveServices:     statusNode.InactiveServices,
		EventCount:           statusNode.LogEventCount,
		AnalyzedCount:        statusNode.LogEventAnalyzedCount,
		PendingPolicyCount:   statusNode.PolicyPendingCount,
		ApprovedPolicyCount:  statusNode.PolicyApprovedCount,
		DismissedPolicyCount: statusNode.PolicyDismissedCount,
	}

	s.scope.Debug("fetched datadog account status",
		log.String("health", string(result.Health)),
		log.Int("serviceCount", result.ServiceCount),
		log.Int("ok", result.OkServices),
		log.Int("inactive", result.InactiveServices))

	return result, nil
}
