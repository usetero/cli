package tui

import (
	"context"
	"fmt"
	"sync"

	"github.com/usetero/cli/internal/domains/identity"
	"github.com/usetero/cli/internal/domains/tenancy"
	psclient "github.com/usetero/cli/internal/infrastructure/powersync/client"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

type sessionStorage struct {
	env string

	mu             sync.RWMutex
	organizationID tenancy.OrganizationID
}

func newSessionStorage(env string) *sessionStorage {
	return &sessionStorage{env: env}
}

func (s *sessionStorage) SetOrganizationID(organizationID tenancy.OrganizationID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.organizationID = organizationID
}

func (s *sessionStorage) DatabasePath(accountID sqlite.AccountID) (sqlite.DatabasePath, error) {
	s.mu.RLock()
	organizationID := s.organizationID
	s.mu.RUnlock()
	if s.env == "" {
		return "", fmt.Errorf("env is required")
	}
	if organizationID == "" {
		return "", fmt.Errorf("organization id is required")
	}
	storage, err := sqlite.NewDefaultStorage(s.env, string(organizationID))
	if err != nil {
		return "", err
	}
	return storage.DatabasePath(accountID)
}

type syncerTokenSource struct {
	identity *identity.Service
}

func (s syncerTokenSource) GetAccessToken(ctx context.Context) (pssyncer.AccessToken, error) {
	token, err := s.identity.GetAccessToken(ctx)
	if err != nil {
		return "", err
	}
	return pssyncer.AccessToken(token), nil
}

func (s syncerTokenSource) ForceRefreshAccessToken(ctx context.Context) (pssyncer.AccessToken, error) {
	return s.GetAccessToken(ctx)
}

type uploaderTokenSource struct {
	identity *identity.Service
}

func (s uploaderTokenSource) GetAccessToken(ctx context.Context) (psclient.AccessToken, error) {
	token, err := s.identity.GetAccessToken(ctx)
	if err != nil {
		return "", err
	}
	return psclient.AccessToken(token), nil
}
