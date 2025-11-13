package account

// DefaultAccountSaver defines the interface for saving and retrieving default account preferences.
type DefaultAccountSaver interface {
	GetDefaultAccountID() string
	SetDefaultAccountID(accountID string) error
}
