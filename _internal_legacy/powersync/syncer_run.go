package powersync

import (
	"context"
	"errors"
	"fmt"
	"time"

	psapi "github.com/usetero/cli/internal/boundary/powersync"
	"github.com/usetero/cli/internal/log"
)

const (
	initialRetryDelay = 1 * time.Second
	maxRetryDelay     = 10 * time.Second
	errorStateAfter   = 3 // show error state after this many consecutive failures
)

// run is the main sync loop with retry logic.
func (s *syncer) run(ctx context.Context) {
	defer close(s.done)

	retryDelay := initialRetryDelay
	retries := 0

	for {
		if ctx.Err() != nil {
			return
		}

		err := s.runSession(ctx)
		if err == nil {
			retryDelay = initialRetryDelay
			retries = 0
			continue
		}
		if ctx.Err() != nil {
			return
		}

		var clientErr *psapi.Error
		if errors.As(err, &clientErr) {
			if clientErr.IsAuth() {
				s.scope.Debug("auth error, force-refreshing token")
				s.setState(NewReconnecting(false))
				if err := s.forceRefreshToken(ctx); err != nil {
					// Refresh failed — backoff and try again later.
					// Don't give up. The server may be down, we'll recover when it's back.
					retries++
					s.scope.Debug("token refresh failed, retrying", log.Duration("delay", retryDelay), log.Any("error", err))
					s.setState(NewReconnecting(retries >= errorStateAfter))
					s.wait(ctx, retryDelay)
					retryDelay = min(retryDelay*2, maxRetryDelay)
				}
				continue
			}

			if clientErr.IsPermanent() {
				s.setError(err)
				return
			}
		}

		// Transient API errors and non-API errors (e.g. extension state
		// errors) are retried with backoff. runSession calls Start() which
		// resets the extension state via tear_down(), so retry is safe.
		retries++
		s.scope.Debug("transient error, retrying", log.Duration("delay", retryDelay), log.Int("attempt", retries), log.Any("error", err))
		s.setState(NewReconnecting(retries >= errorStateAfter))
		s.wait(ctx, retryDelay)
		retryDelay = min(retryDelay*2, maxRetryDelay)
	}
}

func (s *syncer) refreshToken(ctx context.Context) error {
	token, err := s.tokenRefresher.GetAccessToken(ctx)
	if err != nil {
		return err
	}
	s.client.SetToken(token)
	if _, err := s.controlPlaneNotifyTokenRefreshed(ctx); err != nil {
		return fmt.Errorf("notify token refreshed: %w", err)
	}
	return nil
}

// forceRefreshToken unconditionally refreshes the token, bypassing local
// expiration checks. Used when the server has rejected the current token.
func (s *syncer) forceRefreshToken(ctx context.Context) error {
	token, err := s.tokenRefresher.ForceRefreshAccessToken(ctx)
	if err != nil {
		return err
	}
	s.client.SetToken(token)
	if _, err := s.controlPlaneNotifyTokenRefreshed(ctx); err != nil {
		return fmt.Errorf("notify token refreshed: %w", err)
	}
	return nil
}

func (s *syncer) setError(err error) {
	s.setState(NewError(err))
	s.scope.Error("sync failed", log.Any("error", err))
}

func (s *syncer) wait(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}
