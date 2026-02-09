package preferences

import (
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
)

// Theme represents a color theme preference.
type Theme string

const (
	ThemeAuto  Theme = ""      // auto-detect from terminal
	ThemeDark  Theme = "dark"  // force dark mode
	ThemeLight Theme = "light" // force light mode
)

// Store keys for user preferences.
const (
	keyTheme       = "theme"
	keyActiveOrgID = "active_org_id"
	keyRole        = "role"
)

// UserPreferences provides access to env-level user preferences.
// These persist across org switches.
type UserPreferences interface {
	GetTheme() Theme
	SetTheme(theme Theme) error
	GetActiveOrgID() domain.OrganizationID
	SetActiveOrgID(orgID domain.OrganizationID) error
	GetRole() string
	SetRole(role string) error
	Clear() error
}

// UserService implements UserPreferences backed by an env-level Store.
type UserService struct {
	store Store
	scope log.Scope
}

var _ UserPreferences = (*UserService)(nil)

// NewUserService creates a new user preferences service.
func NewUserService(store Store, scope log.Scope) *UserService {
	return &UserService{
		store: store,
		scope: scope.Child("preferences.user"),
	}
}

func (s *UserService) GetTheme() Theme {
	return Theme(s.store.Get(keyTheme))
}

func (s *UserService) SetTheme(theme Theme) error {
	s.store.Set(keyTheme, string(theme))
	return s.store.Save()
}

func (s *UserService) GetActiveOrgID() domain.OrganizationID {
	return domain.OrganizationID(s.store.Get(keyActiveOrgID))
}

func (s *UserService) SetActiveOrgID(orgID domain.OrganizationID) error {
	s.store.Set(keyActiveOrgID, orgID.String())
	return s.store.Save()
}

func (s *UserService) GetRole() string {
	return s.store.Get(keyRole)
}

func (s *UserService) SetRole(role string) error {
	s.store.Set(keyRole, role)
	return s.store.Save()
}

func (s *UserService) Clear() error {
	return s.store.Clear()
}
