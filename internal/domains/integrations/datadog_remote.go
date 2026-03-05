package integrations

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/domains/tenancy"
	controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"
)

type remoteDatadogClient interface {
	GetAccountDatadogAccount(ctx context.Context, accountID controlplane.AccountID) (*controlplane.DatadogAccount, error)
	ValidateDatadogAPIKey(ctx context.Context, apiKey string, site controlplane.DatadogSite) (bool, string, error)
	CreateDatadogAccountWithCredentials(ctx context.Context, input controlplane.CreateDatadogAccountInput) (controlplane.DatadogAccount, error)
	GetDatadogAccountStatus(ctx context.Context, datadogAccountID controlplane.DatadogAccountID) (*controlplane.DatadogAccountStatus, error)
}

// RemoteDatadogService uses control-plane API for Datadog onboarding operations.
type RemoteDatadogService struct {
	client remoteDatadogClient
}

func NewRemoteDatadogService(client remoteDatadogClient) *RemoteDatadogService {
	return &RemoteDatadogService{client: client}
}

func (s *RemoteDatadogService) GetByAccount(ctx context.Context, accountID tenancy.AccountID) (*DatadogAccount, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("integrations remote datadog service is not initialized")
	}
	if accountID == "" {
		return nil, fmt.Errorf("account id is required")
	}

	account, err := s.client.GetAccountDatadogAccount(ctx, toControlPlaneAccountID(accountID))
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, nil
	}
	mapped := fromControlPlaneDatadogAccount(*account)
	return &mapped, nil
}

func (s *RemoteDatadogService) ValidateAPIKey(ctx context.Context, site DatadogSite, apiKey string) (bool, string, error) {
	if s == nil || s.client == nil {
		return false, "", fmt.Errorf("integrations remote datadog service is not initialized")
	}
	if !site.Valid() {
		return false, "", fmt.Errorf("datadog site is required")
	}
	if apiKey == "" {
		return false, "", fmt.Errorf("api key is required")
	}
	return s.client.ValidateDatadogAPIKey(ctx, apiKey, toControlPlaneDatadogSite(site))
}

func (s *RemoteDatadogService) Create(ctx context.Context, input CreateDatadogAccountInput) (DatadogAccountID, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("integrations remote datadog service is not initialized")
	}
	if input.AccountID == "" {
		return "", fmt.Errorf("account id is required")
	}
	if input.Name == "" {
		return "", fmt.Errorf("name is required")
	}
	if !input.Site.Valid() {
		return "", fmt.Errorf("datadog site is required")
	}
	if input.APIKey == "" {
		return "", fmt.Errorf("api key is required")
	}
	if input.AppKey == "" {
		return "", fmt.Errorf("app key is required")
	}

	created, err := s.client.CreateDatadogAccountWithCredentials(ctx, toControlPlaneCreateDatadogAccountInput(input))
	if err != nil {
		return "", err
	}
	return fromControlPlaneDatadogAccount(created).ID, nil
}

func (s *RemoteDatadogService) Status(ctx context.Context, datadogAccountID DatadogAccountID) (*DatadogStatus, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("integrations remote datadog service is not initialized")
	}
	if datadogAccountID == "" {
		return nil, fmt.Errorf("datadog account id is required")
	}

	status, err := s.client.GetDatadogAccountStatus(ctx, toControlPlaneDatadogAccountID(datadogAccountID))
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, nil
	}
	mapped := fromControlPlaneDatadogStatus(*status)
	return &mapped, nil
}
