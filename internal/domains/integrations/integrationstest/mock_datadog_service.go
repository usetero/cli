package integrationstest

import (
	"context"

	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/tenancy"
)

// MockDatadogService is a functional mock for integrations.DatadogService.
type MockDatadogService struct {
	GetByAccountFn   func(ctx context.Context, accountID tenancy.AccountID) (*integrations.DatadogAccount, error)
	ValidateAPIKeyFn func(ctx context.Context, validation integrations.DatadogAPIKeyValidation) (bool, string, error)
	CreateFn         func(ctx context.Context, create integrations.DatadogAccountCreate) (integrations.DatadogAccountID, error)
	StatusFn         func(ctx context.Context, datadogAccountID integrations.DatadogAccountID) (*integrations.DatadogStatus, error)
}

var _ integrations.DatadogService = (*MockDatadogService)(nil)

func NewMockDatadogService() *MockDatadogService {
	return &MockDatadogService{}
}

func (m *MockDatadogService) GetByAccount(ctx context.Context, accountID tenancy.AccountID) (*integrations.DatadogAccount, error) {
	if m.GetByAccountFn == nil {
		return nil, nil
	}
	return m.GetByAccountFn(ctx, accountID)
}

func (m *MockDatadogService) ValidateAPIKey(ctx context.Context, validation integrations.DatadogAPIKeyValidation) (bool, string, error) {
	if m.ValidateAPIKeyFn == nil {
		return false, "", nil
	}
	return m.ValidateAPIKeyFn(ctx, validation)
}

func (m *MockDatadogService) Create(ctx context.Context, create integrations.DatadogAccountCreate) (integrations.DatadogAccountID, error) {
	if m.CreateFn == nil {
		return "", nil
	}
	return m.CreateFn(ctx, create)
}

func (m *MockDatadogService) Status(ctx context.Context, datadogAccountID integrations.DatadogAccountID) (*integrations.DatadogStatus, error) {
	if m.StatusFn == nil {
		return nil, nil
	}
	return m.StatusFn(ctx, datadogAccountID)
}
