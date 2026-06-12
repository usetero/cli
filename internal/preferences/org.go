package preferences

import (
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
)

// Store keys for org preferences.
const (
	keyDefaultAccountID = "default_account_id"
)

// OrgPreferences provides access to org-scoped preferences.
// These are swapped when the user switches organizations.
type OrgPreferences interface {
	GetDefaultAccountID() domain.AccountID
	SetDefaultAccountID(accountID domain.AccountID) error
	ClearDefaultAccountID() error
	Clear() error
}

// OrgService implements OrgPreferences backed by an org-level Store.
type OrgService struct {
	store Store
	scope log.Scope
}

var _ OrgPreferences = (*OrgService)(nil)

// NewOrgService creates a new org preferences service.
func NewOrgService(store Store, scope log.Scope) *OrgService {
	return &OrgService{
		store: store,
		scope: scope.Child("preferences.org"),
	}
}

func (s *OrgService) GetDefaultAccountID() domain.AccountID {
	return domain.AccountID(s.store.Get(keyDefaultAccountID))
}

func (s *OrgService) SetDefaultAccountID(accountID domain.AccountID) error {
	s.store.Set(keyDefaultAccountID, accountID.String())
	return s.store.Save()
}

// ClearDefaultAccountID clears the default account.
func (s *OrgService) ClearDefaultAccountID() error {
	s.store.Set(keyDefaultAccountID, "")
	return s.store.Save()
}

func (s *OrgService) Clear() error {
	return s.store.Clear()
}
