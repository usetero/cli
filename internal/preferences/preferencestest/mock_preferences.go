package preferencestest

import (
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/preferences"
)

// MockUserPreferences implements preferences.UserPreferences for testing.
type MockUserPreferences struct {
	GetThemeFunc       func() preferences.Theme
	SetThemeFunc       func(theme preferences.Theme) error
	GetActiveOrgIDFunc func() domain.OrganizationID
	SetActiveOrgIDFunc func(orgID domain.OrganizationID) error
	GetRoleFunc        func() string
	SetRoleFunc        func(role string) error
	ClearFunc          func() error
}

func NewMockUserPreferences() *MockUserPreferences {
	return &MockUserPreferences{}
}

func (m *MockUserPreferences) GetTheme() preferences.Theme {
	if m.GetThemeFunc != nil {
		return m.GetThemeFunc()
	}
	return ""
}

func (m *MockUserPreferences) SetTheme(theme preferences.Theme) error {
	if m.SetThemeFunc != nil {
		return m.SetThemeFunc(theme)
	}
	return nil
}

func (m *MockUserPreferences) GetActiveOrgID() domain.OrganizationID {
	if m.GetActiveOrgIDFunc != nil {
		return m.GetActiveOrgIDFunc()
	}
	return ""
}

func (m *MockUserPreferences) SetActiveOrgID(orgID domain.OrganizationID) error {
	if m.SetActiveOrgIDFunc != nil {
		return m.SetActiveOrgIDFunc(orgID)
	}
	return nil
}

func (m *MockUserPreferences) GetRole() string {
	if m.GetRoleFunc != nil {
		return m.GetRoleFunc()
	}
	return ""
}

func (m *MockUserPreferences) SetRole(role string) error {
	if m.SetRoleFunc != nil {
		return m.SetRoleFunc(role)
	}
	return nil
}

func (m *MockUserPreferences) Clear() error {
	if m.ClearFunc != nil {
		return m.ClearFunc()
	}
	return nil
}

// MockOrgPreferences implements preferences.OrgPreferences for testing.
type MockOrgPreferences struct {
	GetDefaultAccountIDFunc   func() domain.AccountID
	SetDefaultAccountIDFunc   func(accountID domain.AccountID) error
	ClearDefaultAccountIDFunc func() error
	ClearFunc                 func() error
}

func NewMockOrgPreferences() *MockOrgPreferences {
	return &MockOrgPreferences{}
}

func (m *MockOrgPreferences) GetDefaultAccountID() domain.AccountID {
	if m.GetDefaultAccountIDFunc != nil {
		return m.GetDefaultAccountIDFunc()
	}
	return ""
}

func (m *MockOrgPreferences) SetDefaultAccountID(accountID domain.AccountID) error {
	if m.SetDefaultAccountIDFunc != nil {
		return m.SetDefaultAccountIDFunc(accountID)
	}
	return nil
}

func (m *MockOrgPreferences) ClearDefaultAccountID() error {
	if m.ClearDefaultAccountIDFunc != nil {
		return m.ClearDefaultAccountIDFunc()
	}
	return nil
}

func (m *MockOrgPreferences) Clear() error {
	if m.ClearFunc != nil {
		return m.ClearFunc()
	}
	return nil
}
