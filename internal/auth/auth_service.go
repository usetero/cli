package auth

import (
	"context"
	"errors"
	"time"

	"github.com/usetero/cli/internal/log"
)

// Service handles authentication business logic.
// It coordinates between the OAuth provider and secure token storage.
// It defines domain concepts (access_token, refresh_token) and translates them
// to/from generic key-value storage operations.
type Service struct {
	provider OAuthProvider
	storage  SecureStorage
	logger   log.Logger
}

// NewService creates a new authentication service.
func NewService(provider OAuthProvider, storage SecureStorage, logger log.Logger) *Service {
	return &Service{
		provider: provider,
		storage:  storage,
		logger:   logger,
	}
}

// DeviceAuth contains the information needed to display to the user.
type DeviceAuth struct {
	UserCode                string
	VerificationURI         string
	VerificationURIComplete string
	DeviceCode              string // Kept internally for polling
	ExpiresIn               int
	Interval                int
}

// Result contains the tokens and user information after successful authentication.
type Result struct {
	AccessToken  string
	RefreshToken string
	User         User
}

// User represents an authenticated user.
type User struct {
	ID            string
	Email         string
	EmailVerified bool
	FirstName     string
	LastName      string
}

// StartDeviceAuth initiates the device authorization flow.
func (s *Service) StartDeviceAuth(ctx context.Context) (*DeviceAuth, error) {
	s.logger.Debug("starting device authorization flow")

	resp, err := s.provider.AuthorizeDevice(ctx)
	if err != nil {
		s.logger.Error("failed to start device authorization", "error", err)
		return nil, err
	}

	s.logger.Debug("device authorization started",
		"user_code", resp.UserCode,
		"expires_in", resp.ExpiresIn,
		"interval", resp.Interval)

	return &DeviceAuth{
		UserCode:                resp.UserCode,
		VerificationURI:         resp.VerificationURI,
		VerificationURIComplete: resp.VerificationURIComplete,
		DeviceCode:              resp.DeviceCode,
		ExpiresIn:               resp.ExpiresIn,
		Interval:                resp.Interval,
	}, nil
}

// WaitForAuth polls the OAuth provider until the user completes authentication or an error occurs.
// This is a blocking call that handles the polling loop with proper backoff.
func (s *Service) WaitForAuth(ctx context.Context, deviceCode string, interval time.Duration) (*Result, error) {
	s.logger.Debug("starting authentication polling", "interval", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	currentInterval := interval

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			resp, err := s.provider.PollAuthentication(ctx, deviceCode)
			if err == nil {
				// Success! Save tokens and return
				s.logger.Info("authentication successful", "user_id", resp.User.ID, "email", resp.User.Email)

				if err := s.saveTokens(resp.AccessToken, resp.RefreshToken); err != nil {
					s.logger.Error("failed to save tokens", "error", err)
					return nil, err
				}

				return &Result{
					AccessToken:  resp.AccessToken,
					RefreshToken: resp.RefreshToken,
					User:         resp.User,
				}, nil
			}

			// Handle specific error types
			var pendingErr *AuthorizationPendingError
			var slowDownErr *SlowDownError
			var expiredErr *ExpiredTokenError
			var deniedErr *AccessDeniedError

			switch {
			case errors.As(err, &pendingErr):
				// Still waiting - continue polling
				s.logger.Debug("authorization pending, continuing to poll")
				continue

			case errors.As(err, &slowDownErr):
				// Increase polling interval
				currentInterval = currentInterval * 2
				ticker.Reset(currentInterval)
				s.logger.Debug("slowing down polling", "new_interval", currentInterval)
				continue

			case errors.As(err, &expiredErr):
				s.logger.Error("device code expired")
				return nil, errors.New("device code expired - please restart authentication")

			case errors.As(err, &deniedErr):
				s.logger.Info("user denied authorization")
				return nil, errors.New("user denied authorization")

			default:
				// Unknown error
				s.logger.Error("authentication polling failed", "error", err)
				return nil, err
			}
		}
	}
}

// IsAuthenticated checks if the user has valid stored credentials.
func (s *Service) IsAuthenticated() bool {
	accessToken, err := s.storage.Get("access_token")
	if err != nil {
		s.logger.Error("failed to check authentication", "error", err)
		return false
	}
	return accessToken != ""
}

// GetAccessToken retrieves the stored access token.
func (s *Service) GetAccessToken(ctx context.Context) (string, error) {
	accessToken, err := s.storage.Get("access_token")
	if err != nil {
		s.logger.Error("failed to get access token", "error", err)
		return "", err
	}
	if accessToken == "" {
		return "", errors.New("no access token found")
	}
	return accessToken, nil
}

// ClearTokens removes all stored authentication tokens.
func (s *Service) ClearTokens() error {
	s.logger.Info("clearing authentication tokens")
	if err := s.storage.Delete("access_token"); err != nil {
		s.logger.Error("failed to delete access token", "error", err)
		return err
	}
	if err := s.storage.Delete("refresh_token"); err != nil {
		s.logger.Error("failed to delete refresh token", "error", err)
		return err
	}
	return nil
}

// saveTokens stores the access and refresh tokens securely.
func (s *Service) saveTokens(accessToken, refreshToken string) error {
	if err := s.storage.Set("access_token", accessToken); err != nil {
		return err
	}
	if err := s.storage.Set("refresh_token", refreshToken); err != nil {
		return err
	}
	return nil
}
