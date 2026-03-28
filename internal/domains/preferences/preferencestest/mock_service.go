package preferencestest

import (
	"context"

	"github.com/usetero/cli/internal/domains/preferences"
)

// MockService is a functional mock for preferences.PreferenceService.
type MockService struct {
	SnapshotFn        func(ctx context.Context) (preferences.Snapshot, error)
	SetRoleFn         func(ctx context.Context, selection preferences.RoleSelection) error
	SetOrganizationFn func(ctx context.Context, selection preferences.OrganizationSelection) error
	SetAccountFn      func(ctx context.Context, selection preferences.AccountSelection) error
	SetWorkspaceFn    func(ctx context.Context, selection preferences.WorkspaceSelection) error
	SetScopeFn        func(ctx context.Context, selection preferences.ScopeSelection) error
	ClearScopeFn      func(ctx context.Context) error
}

var _ preferences.PreferenceService = (*MockService)(nil)

func NewMockService() *MockService {
	return &MockService{}
}

func (m *MockService) Snapshot(ctx context.Context) (preferences.Snapshot, error) {
	if m.SnapshotFn == nil {
		return preferences.Snapshot{}, nil
	}
	return m.SnapshotFn(ctx)
}

func (m *MockService) SetRole(ctx context.Context, selection preferences.RoleSelection) error {
	if m.SetRoleFn == nil {
		return nil
	}
	return m.SetRoleFn(ctx, selection)
}

func (m *MockService) SetOrganization(ctx context.Context, selection preferences.OrganizationSelection) error {
	if m.SetOrganizationFn == nil {
		return nil
	}
	return m.SetOrganizationFn(ctx, selection)
}

func (m *MockService) SetAccount(ctx context.Context, selection preferences.AccountSelection) error {
	if m.SetAccountFn == nil {
		return nil
	}
	return m.SetAccountFn(ctx, selection)
}

func (m *MockService) SetWorkspace(ctx context.Context, selection preferences.WorkspaceSelection) error {
	if m.SetWorkspaceFn == nil {
		return nil
	}
	return m.SetWorkspaceFn(ctx, selection)
}

func (m *MockService) SetScope(ctx context.Context, selection preferences.ScopeSelection) error {
	if m.SetScopeFn == nil {
		return nil
	}
	return m.SetScopeFn(ctx, selection)
}

func (m *MockService) ClearScope(ctx context.Context) error {
	if m.ClearScopeFn == nil {
		return nil
	}
	return m.ClearScopeFn(ctx)
}
