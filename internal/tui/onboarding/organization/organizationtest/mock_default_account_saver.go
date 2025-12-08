package organizationtest

// MockDefaultAccountSaver implements organization.DefaultAccountSaver for testing.
type MockDefaultAccountSaver struct {
	GetDefaultAccountIDFunc func() string
	SetDefaultAccountIDFunc func(accountID string) error
}

func (m *MockDefaultAccountSaver) GetDefaultAccountID() string {
	if m.GetDefaultAccountIDFunc != nil {
		return m.GetDefaultAccountIDFunc()
	}
	return ""
}

func (m *MockDefaultAccountSaver) SetDefaultAccountID(accountID string) error {
	if m.SetDefaultAccountIDFunc != nil {
		return m.SetDefaultAccountIDFunc(accountID)
	}
	return nil
}
