package onboardingtest

import (
	"context"

	"github.com/usetero/cli/internal/domains/tenancy"
)

type WorkspaceService struct {
	ListByAccountValue map[tenancy.AccountID][]tenancy.Workspace
}

func (s *WorkspaceService) Create(context.Context, tenancy.AccountID, string) (tenancy.WorkspaceID, error) {
	return "", nil
}
func (s *WorkspaceService) Delete(context.Context, tenancy.WorkspaceID) error { return nil }
func (s *WorkspaceService) ListByAccount(_ context.Context, accountID tenancy.AccountID) ([]tenancy.Workspace, error) {
	return s.ListByAccountValue[accountID], nil
}
