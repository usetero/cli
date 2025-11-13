package preferences

import (
	"github.com/usetero/cli/internal/log"
)

// Service handles user preferences business logic.
// It defines domain concepts (email, orgID, etc.) and translates them
// to/from generic key-value storage operations.
type Service struct {
	store  Store
	logger log.Logger
}

// NewService creates a new preferences service.
func NewService(store Store, logger log.Logger) *Service {
	return &Service{
		store:  store,
		logger: logger,
	}
}

// GetEmail returns the user's email
func (s *Service) GetEmail() string {
	return s.store.Get("email")
}

// SetEmail saves the user's email
func (s *Service) SetEmail(email string) error {
	s.store.Set("email", email)
	return s.store.Save()
}

// GetDatadogAPIKey returns the Datadog API key
func (s *Service) GetDatadogAPIKey() string {
	return s.store.Get("datadog_api_key")
}

// SetDatadogAPIKey saves the Datadog API key
func (s *Service) SetDatadogAPIKey(key string) error {
	s.store.Set("datadog_api_key", key)
	return s.store.Save()
}

// GetDefaultOrgID returns the default organization ID
func (s *Service) GetDefaultOrgID() string {
	return s.store.Get("default_org_id")
}

// SetDefaultOrgID saves the default organization ID
func (s *Service) SetDefaultOrgID(orgID string) error {
	s.store.Set("default_org_id", orgID)
	return s.store.Save()
}

// GetDefaultAccountID returns the default account ID
func (s *Service) GetDefaultAccountID() string {
	return s.store.Get("default_account_id")
}

// SetDefaultAccountID saves the default account ID
func (s *Service) SetDefaultAccountID(accountID string) error {
	s.store.Set("default_account_id", accountID)
	return s.store.Save()
}

// GetDefaultWorkspaceID returns the default workspace ID
func (s *Service) GetDefaultWorkspaceID() string {
	return s.store.Get("default_workspace_id")
}

// SetDefaultWorkspaceID saves the default workspace ID
func (s *Service) SetDefaultWorkspaceID(workspaceID string) error {
	s.store.Set("default_workspace_id", workspaceID)
	return s.store.Save()
}

// ClearEmail clears the user's email (for going back in onboarding)
func (s *Service) ClearEmail() error {
	s.store.Set("email", "")
	return s.store.Save()
}

// ClearDatadogAPIKey clears the Datadog API key (for going back in onboarding)
func (s *Service) ClearDatadogAPIKey() error {
	s.store.Set("datadog_api_key", "")
	return s.store.Save()
}

// ClearDefaultOrgID clears the default organization ID (for going back in onboarding)
func (s *Service) ClearDefaultOrgID() error {
	s.store.Set("default_org_id", "")
	return s.store.Save()
}

// GetHasSeenGreeting returns whether the user has seen the greeting
func (s *Service) GetHasSeenGreeting() bool {
	return s.store.GetBool("has_seen_greeting")
}

// SetHasSeenGreeting saves whether the user has seen the greeting
func (s *Service) SetHasSeenGreeting(seen bool) error {
	s.store.SetBool("has_seen_greeting", seen)
	return s.store.Save()
}

// GetRole returns the user's role in this org
func (s *Service) GetRole() string {
	return s.store.Get("role")
}

// SetRole saves the user's role in this org
func (s *Service) SetRole(role string) error {
	s.store.Set("role", role)
	return s.store.Save()
}

// GetServices returns the services the user owns (if engineer role)
func (s *Service) GetServices() []string {
	return s.store.GetList("services")
}

// SetServices saves the services the user owns (if engineer role)
func (s *Service) SetServices(services []string) error {
	s.store.SetList("services", services)
	return s.store.Save()
}

// ClearRole clears the user's role (for going back in onboarding)
func (s *Service) ClearRole() error {
	s.store.Set("role", "")
	return s.store.Save()
}

// ClearServices clears the user's services (for going back in onboarding)
func (s *Service) ClearServices() error {
	s.store.SetList("services", nil)
	return s.store.Save()
}
