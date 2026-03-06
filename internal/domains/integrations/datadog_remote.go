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
	if client == nil {
		panic("integrations remote datadog service requires client")
	}
	return &RemoteDatadogService{client: client}
}

func (s *RemoteDatadogService) GetByAccount(ctx context.Context, accountID tenancy.AccountID) (*DatadogAccount, error) {
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

func (s *RemoteDatadogService) ValidateAPIKey(ctx context.Context, validation DatadogAPIKeyValidation) (bool, string, error) {
	validated, err := validation.Validate()
	if err != nil {
		return false, "", err
	}
	return s.client.ValidateDatadogAPIKey(ctx, validated.APIKey.String(), toControlPlaneDatadogSite(validated.Site))
}

func (s *RemoteDatadogService) Create(ctx context.Context, create DatadogAccountCreate) (DatadogAccountID, error) {
	validated, err := create.Validate()
	if err != nil {
		return "", err
	}

	created, err := s.client.CreateDatadogAccountWithCredentials(ctx, toControlPlaneCreateDatadogAccountInput(validated))
	if err != nil {
		return "", err
	}
	return fromControlPlaneDatadogAccount(created).ID, nil
}

func (s *RemoteDatadogService) Status(ctx context.Context, datadogAccountID DatadogAccountID) (*DatadogStatus, error) {
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
