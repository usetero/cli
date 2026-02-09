package preferences

import (
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
)

// Store keys.
const (
	keyEmail              = "email"
	keyDatadogAPIKey      = "datadog_api_key"
	keyDefaultOrgID       = "default_org_id"
	keyDefaultOrgName     = "default_org_name"
	keyDefaultAccountID   = "default_account_id"
	keyDefaultWorkspaceID = "default_workspace_id"
	keyHasSeenGreeting    = "has_seen_greeting"
	keyRole               = "role"
	keyServices           = "services"
)

// Preferences provides access to user preferences.
type Preferences interface {
	GetEmail() string
	SetEmail(email string) error
	GetDatadogAPIKey() string
	SetDatadogAPIKey(key string) error
	GetDefaultOrgID() domain.OrganizationID
	SetDefaultOrgID(orgID domain.OrganizationID) error
	GetDefaultOrgName() string
	SetDefaultOrgName(orgName string) error
	GetDefaultAccountID() domain.AccountID
	SetDefaultAccountID(accountID domain.AccountID) error
	GetDefaultWorkspaceID() domain.WorkspaceID
	SetDefaultWorkspaceID(workspaceID domain.WorkspaceID) error
	ClearEmail() error
	ClearDatadogAPIKey() error
	ClearDefaultOrgID() error
	ClearDefaultAccountID() error
	ClearDefaultWorkspaceID() error
	GetHasSeenGreeting() bool
	SetHasSeenGreeting(seen bool) error
	GetRole() string
	SetRole(role string) error
	GetServices() []string
	SetServices(services []string) error
	ClearRole() error
	ClearServices() error
}

// Service handles user preferences business logic.
// It defines domain concepts (email, orgID, etc.) and translates them
// to/from generic key-value storage operations.
type Service struct {
	store Store
	scope log.Scope
}

// Ensure Service implements Preferences.
var _ Preferences = (*Service)(nil)

// NewService creates a new preferences service.
func NewService(store Store, scope log.Scope) *Service {
	return &Service{
		store: store,
		scope: scope.Child("preferences"),
	}
}

func (s *Service) GetEmail() string {
	return s.store.Get(keyEmail)
}

func (s *Service) SetEmail(email string) error {
	s.store.Set(keyEmail, email)
	return s.store.Save()
}

func (s *Service) GetDatadogAPIKey() string {
	return s.store.Get(keyDatadogAPIKey)
}

func (s *Service) SetDatadogAPIKey(key string) error {
	s.store.Set(keyDatadogAPIKey, key)
	return s.store.Save()
}

func (s *Service) GetDefaultOrgID() domain.OrganizationID {
	return domain.OrganizationID(s.store.Get(keyDefaultOrgID))
}

func (s *Service) SetDefaultOrgID(orgID domain.OrganizationID) error {
	s.store.Set(keyDefaultOrgID, orgID.String())
	return s.store.Save()
}

func (s *Service) GetDefaultOrgName() string {
	return s.store.Get(keyDefaultOrgName)
}

func (s *Service) SetDefaultOrgName(orgName string) error {
	s.store.Set(keyDefaultOrgName, orgName)
	return s.store.Save()
}

func (s *Service) GetDefaultAccountID() domain.AccountID {
	return domain.AccountID(s.store.Get(keyDefaultAccountID))
}

func (s *Service) SetDefaultAccountID(accountID domain.AccountID) error {
	s.store.Set(keyDefaultAccountID, accountID.String())
	return s.store.Save()
}

func (s *Service) GetDefaultWorkspaceID() domain.WorkspaceID {
	return domain.WorkspaceID(s.store.Get(keyDefaultWorkspaceID))
}

func (s *Service) SetDefaultWorkspaceID(workspaceID domain.WorkspaceID) error {
	s.store.Set(keyDefaultWorkspaceID, workspaceID.String())
	return s.store.Save()
}

func (s *Service) ClearEmail() error {
	s.store.Set(keyEmail, "")
	return s.store.Save()
}

func (s *Service) ClearDatadogAPIKey() error {
	s.store.Set(keyDatadogAPIKey, "")
	return s.store.Save()
}

// ClearDefaultOrgID clears the default organization ID and cascades to
// account and workspace since they are org-scoped.
func (s *Service) ClearDefaultOrgID() error {
	s.store.Set(keyDefaultOrgID, "")
	s.store.Set(keyDefaultOrgName, "")
	s.store.Set(keyDefaultAccountID, "")
	s.store.Set(keyDefaultWorkspaceID, "")
	return s.store.Save()
}

// ClearDefaultAccountID clears the default account ID and cascades to
// workspace since it is account-scoped.
func (s *Service) ClearDefaultAccountID() error {
	s.store.Set(keyDefaultAccountID, "")
	s.store.Set(keyDefaultWorkspaceID, "")
	return s.store.Save()
}

// ClearDefaultWorkspaceID clears the default workspace ID.
func (s *Service) ClearDefaultWorkspaceID() error {
	s.store.Set(keyDefaultWorkspaceID, "")
	return s.store.Save()
}

func (s *Service) GetHasSeenGreeting() bool {
	return s.store.GetBool(keyHasSeenGreeting)
}

func (s *Service) SetHasSeenGreeting(seen bool) error {
	s.store.SetBool(keyHasSeenGreeting, seen)
	return s.store.Save()
}

func (s *Service) GetRole() string {
	return s.store.Get(keyRole)
}

func (s *Service) SetRole(role string) error {
	s.store.Set(keyRole, role)
	return s.store.Save()
}

func (s *Service) GetServices() []string {
	return s.store.GetList(keyServices)
}

func (s *Service) SetServices(services []string) error {
	s.store.SetList(keyServices, services)
	return s.store.Save()
}

func (s *Service) ClearRole() error {
	s.store.Set(keyRole, "")
	return s.store.Save()
}

func (s *Service) ClearServices() error {
	s.store.SetList(keyServices, nil)
	return s.store.Save()
}

// Clear removes all preferences.
func (s *Service) Clear() error {
	return s.store.Clear()
}
