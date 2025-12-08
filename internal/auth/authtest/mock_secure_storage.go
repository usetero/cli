package authtest

// MockSecureStorage is a test double for auth.SecureStorage.
type MockSecureStorage struct {
	GetFunc    func(key string) (string, error)
	SetFunc    func(key string, value string) error
	DeleteFunc func(key string) error
}

func (m *MockSecureStorage) Get(key string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(key)
	}
	return "", nil
}

func (m *MockSecureStorage) Set(key string, value string) error {
	if m.SetFunc != nil {
		return m.SetFunc(key, value)
	}
	return nil
}

func (m *MockSecureStorage) Delete(key string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(key)
	}
	return nil
}
