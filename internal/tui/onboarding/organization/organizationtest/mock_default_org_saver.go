package organizationtest

// MockDefaultOrgSaver implements organization.DefaultOrgSaver for testing.
type MockDefaultOrgSaver struct {
	GetDefaultOrgIDFunc func() string
	SetDefaultOrgIDFunc func(orgID string) error
}

func (m *MockDefaultOrgSaver) GetDefaultOrgID() string {
	if m.GetDefaultOrgIDFunc != nil {
		return m.GetDefaultOrgIDFunc()
	}
	return ""
}

func (m *MockDefaultOrgSaver) SetDefaultOrgID(orgID string) error {
	if m.SetDefaultOrgIDFunc != nil {
		return m.SetDefaultOrgIDFunc(orgID)
	}
	return nil
}
