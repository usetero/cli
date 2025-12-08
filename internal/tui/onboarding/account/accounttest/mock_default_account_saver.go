package accounttest

// MockDefaultAccountSaver is a mock implementation of account.DefaultAccountSaver.
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
