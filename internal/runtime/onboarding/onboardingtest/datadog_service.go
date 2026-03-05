package onboardingtest

import (
	"context"

	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/domains/tenancy"
)

type DatadogService struct {
	ByAccountValue map[tenancy.AccountID]*integrations.DatadogAccount
	StatusValue    map[integrations.DatadogAccountID]*integrations.DatadogStatus
}

func (s *DatadogService) GetByAccount(_ context.Context, accountID tenancy.AccountID) (*integrations.DatadogAccount, error) {
	return s.ByAccountValue[accountID], nil
}
func (s *DatadogService) ValidateAPIKey(context.Context, integrations.DatadogSite, string) (bool, string, error) {
	return true, "", nil
}
func (s *DatadogService) Create(context.Context, integrations.CreateDatadogAccountInput) (integrations.DatadogAccountID, error) {
	return "dd_1", nil
}
func (s *DatadogService) Status(_ context.Context, id integrations.DatadogAccountID) (*integrations.DatadogStatus, error) {
	return s.StatusValue[id], nil
}
