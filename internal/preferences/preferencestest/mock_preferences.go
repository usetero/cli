package preferencestest

// MockPreferences implements preferences.Preferences for testing.
type MockPreferences struct {
	GetEmailFunc              func() string
	SetEmailFunc              func(email string) error
	GetDatadogAPIKeyFunc      func() string
	SetDatadogAPIKeyFunc      func(key string) error
	GetDefaultOrgIDFunc       func() string
	SetDefaultOrgIDFunc       func(orgID string) error
	GetDefaultOrgNameFunc     func() string
	SetDefaultOrgNameFunc     func(orgName string) error
	GetDefaultAccountIDFunc   func() string
	SetDefaultAccountIDFunc   func(accountID string) error
	GetDefaultWorkspaceIDFunc func() string
	SetDefaultWorkspaceIDFunc func(workspaceID string) error
	ClearEmailFunc            func() error
	ClearDatadogAPIKeyFunc    func() error
	ClearDefaultOrgIDFunc     func() error
	GetHasSeenGreetingFunc    func() bool
	SetHasSeenGreetingFunc    func(seen bool) error
	GetRoleFunc               func() string
	SetRoleFunc               func(role string) error
	GetServicesFunc           func() []string
	SetServicesFunc           func(services []string) error
	ClearRoleFunc             func() error
	ClearServicesFunc         func() error
}

func (m *MockPreferences) GetEmail() string {
	if m.GetEmailFunc != nil {
		return m.GetEmailFunc()
	}
	return ""
}

func (m *MockPreferences) SetEmail(email string) error {
	if m.SetEmailFunc != nil {
		return m.SetEmailFunc(email)
	}
	return nil
}

func (m *MockPreferences) GetDatadogAPIKey() string {
	if m.GetDatadogAPIKeyFunc != nil {
		return m.GetDatadogAPIKeyFunc()
	}
	return ""
}

func (m *MockPreferences) SetDatadogAPIKey(key string) error {
	if m.SetDatadogAPIKeyFunc != nil {
		return m.SetDatadogAPIKeyFunc(key)
	}
	return nil
}

func (m *MockPreferences) GetDefaultOrgID() string {
	if m.GetDefaultOrgIDFunc != nil {
		return m.GetDefaultOrgIDFunc()
	}
	return ""
}

func (m *MockPreferences) SetDefaultOrgID(orgID string) error {
	if m.SetDefaultOrgIDFunc != nil {
		return m.SetDefaultOrgIDFunc(orgID)
	}
	return nil
}

func (m *MockPreferences) GetDefaultOrgName() string {
	if m.GetDefaultOrgNameFunc != nil {
		return m.GetDefaultOrgNameFunc()
	}
	return ""
}

func (m *MockPreferences) SetDefaultOrgName(orgName string) error {
	if m.SetDefaultOrgNameFunc != nil {
		return m.SetDefaultOrgNameFunc(orgName)
	}
	return nil
}

func (m *MockPreferences) GetDefaultAccountID() string {
	if m.GetDefaultAccountIDFunc != nil {
		return m.GetDefaultAccountIDFunc()
	}
	return ""
}

func (m *MockPreferences) SetDefaultAccountID(accountID string) error {
	if m.SetDefaultAccountIDFunc != nil {
		return m.SetDefaultAccountIDFunc(accountID)
	}
	return nil
}

func (m *MockPreferences) GetDefaultWorkspaceID() string {
	if m.GetDefaultWorkspaceIDFunc != nil {
		return m.GetDefaultWorkspaceIDFunc()
	}
	return ""
}

func (m *MockPreferences) SetDefaultWorkspaceID(workspaceID string) error {
	if m.SetDefaultWorkspaceIDFunc != nil {
		return m.SetDefaultWorkspaceIDFunc(workspaceID)
	}
	return nil
}

func (m *MockPreferences) ClearEmail() error {
	if m.ClearEmailFunc != nil {
		return m.ClearEmailFunc()
	}
	return nil
}

func (m *MockPreferences) ClearDatadogAPIKey() error {
	if m.ClearDatadogAPIKeyFunc != nil {
		return m.ClearDatadogAPIKeyFunc()
	}
	return nil
}

func (m *MockPreferences) ClearDefaultOrgID() error {
	if m.ClearDefaultOrgIDFunc != nil {
		return m.ClearDefaultOrgIDFunc()
	}
	return nil
}

func (m *MockPreferences) GetHasSeenGreeting() bool {
	if m.GetHasSeenGreetingFunc != nil {
		return m.GetHasSeenGreetingFunc()
	}
	return false
}

func (m *MockPreferences) SetHasSeenGreeting(seen bool) error {
	if m.SetHasSeenGreetingFunc != nil {
		return m.SetHasSeenGreetingFunc(seen)
	}
	return nil
}

func (m *MockPreferences) GetRole() string {
	if m.GetRoleFunc != nil {
		return m.GetRoleFunc()
	}
	return ""
}

func (m *MockPreferences) SetRole(role string) error {
	if m.SetRoleFunc != nil {
		return m.SetRoleFunc(role)
	}
	return nil
}

func (m *MockPreferences) GetServices() []string {
	if m.GetServicesFunc != nil {
		return m.GetServicesFunc()
	}
	return nil
}

func (m *MockPreferences) SetServices(services []string) error {
	if m.SetServicesFunc != nil {
		return m.SetServicesFunc(services)
	}
	return nil
}

func (m *MockPreferences) ClearRole() error {
	if m.ClearRoleFunc != nil {
		return m.ClearRoleFunc()
	}
	return nil
}

func (m *MockPreferences) ClearServices() error {
	if m.ClearServicesFunc != nil {
		return m.ClearServicesFunc()
	}
	return nil
}
