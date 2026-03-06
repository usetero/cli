package syncertest

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
)

type TokenSource struct {
	mu sync.RWMutex

	Token               string
	ForceToken          string
	ForceRefreshStarted chan struct{}
	ForceRefreshGate    <-chan struct{}

	ForceCalls atomic.Int32
}

var _ pssyncer.TokenSource = (*TokenSource)(nil)

func (s *TokenSource) GetAccessToken(context.Context) (pssyncer.AccessToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return pssyncer.AccessToken(s.Token), nil
}

func (s *TokenSource) ForceRefreshAccessToken(context.Context) (pssyncer.AccessToken, error) {
	s.ForceCalls.Add(1)

	s.mu.Lock()
	started := s.ForceRefreshStarted
	s.ForceRefreshStarted = nil
	gate := s.ForceRefreshGate
	s.mu.Unlock()

	if started != nil {
		close(started)
	}
	if gate != nil {
		<-gate
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ForceToken == "" {
		return "", fmt.Errorf("no force token")
	}
	s.Token = s.ForceToken
	return pssyncer.AccessToken(s.ForceToken), nil
}
