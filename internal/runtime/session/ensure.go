package session

import "context"

// Ensure makes sure the runtime is active for the requested scope.
// It starts once for first scope and switches (stop + start) when scope changes.
func (s *Service) Ensure(ctx context.Context, scope Scope) error {
	if err := scope.Validate(); err != nil {
		return err
	}
	if err := validateStartAccountID(string(scope.Account.ID)); err != nil {
		return err
	}

	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	scopedStorage, ok := s.storage.(organizationScopedStorage)
	if !ok {
		return ErrStorageNotScopeAware
	}
	scopedStorage.SetOrganizationID(scope.Organization.ID)

	s.mu.RLock()
	running := s.state.Running
	currentScope := s.scope
	s.mu.RUnlock()

	if running && sameRuntimeIdentity(currentScope, scope) {
		s.mu.Lock()
		s.scope = scope
		s.mu.Unlock()
		return nil
	}
	if running {
		if err := s.stopLocked(); err != nil {
			return err
		}
	}
	if err := s.startLocked(ctx, scope.Account.ID); err != nil {
		return err
	}

	s.mu.Lock()
	s.scope = scope
	s.state.OrganizationID = scope.Organization.ID
	s.mu.Unlock()
	return nil
}

func sameRuntimeIdentity(current Scope, next Scope) bool {
	return current.Organization.ID == next.Organization.ID && current.Account.ID == next.Account.ID
}
