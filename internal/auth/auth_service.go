package auth

import (
	"context"
	"errors"
	"time"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
)

// Auth provides authentication operations.
type Auth interface {
	StartDeviceAuth(ctx context.Context) (*DeviceAuth, error)
	WaitForAuth(ctx context.Context, deviceCode string, interval time.Duration) (*Result, error)
	IsAuthenticated() bool
	GetAccessToken(ctx context.Context) (string, error)
	ClearTokens() error
	RefreshTokenWithoutOrganization(ctx context.Context) (string, error)
	RefreshTokenWithOrganization(ctx context.Context, workosOrgID domain.WorkosOrganizationID) (string, error)
}

// Service handles authentication business logic.
// It coordinates between the OAuth provider and secure token storage.
// It defines domain concepts (access_token, refresh_token) and translates them
// to/from generic key-value storage operations.
type Service struct {
	provider OAuthProvider
	storage  SecureStorage
	logger   log.Logger
}

// Ensure Service implements Auth.
var _ Auth = (*Service)(nil)

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
				return nil, errors.New("device code expired - press 'r' to restart")

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

// GetAccessToken retrieves the stored access token, refreshing if expired.
func (s *Service) GetAccessToken(ctx context.Context) (string, error) {
	accessToken, err := s.storage.Get("access_token")
	if err != nil {
		s.logger.Error("failed to get access token", "error", err)
		return "", err
	}
	if accessToken == "" {
		return "", errors.New("no access token found")
	}

	// Check if token is expired
	if isTokenExpired(accessToken) {
		s.logger.Debug("access token expired, refreshing")
		refreshToken, err := s.storage.Get("refresh_token")
		if err != nil {
			s.logger.Error("failed to get refresh token", "error", err)
			return "", err
		}
		if refreshToken == "" {
			return "", errors.New("no refresh token found")
		}

		resp, err := s.provider.RefreshToken(ctx, refreshToken)
		if err != nil {
			s.logger.Error("failed to refresh token", "error", err)
			return "", err
		}

		if err := s.saveTokens(resp.AccessToken, resp.RefreshToken); err != nil {
			s.logger.Error("failed to save refreshed tokens", "error", err)
			return "", err
		}

		s.logger.Debug("token refreshed successfully")
		return resp.AccessToken, nil
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

// RefreshTokenWithoutOrganization refreshes the access token without any organization scope.
// This is used for bootstrap flows where the user needs a user-scoped token to create an org.
// Returns the new access token so callers can update their API clients.
func (s *Service) RefreshTokenWithoutOrganization(ctx context.Context) (string, error) {
	s.logger.Debug("refreshing token without organization scope")

	refreshToken, err := s.storage.Get("refresh_token")
	if err != nil {
		s.logger.Error("failed to get refresh token", "error", err)
		return "", err
	}
	if refreshToken == "" {
		return "", errors.New("no refresh token found")
	}

	resp, err := s.provider.RefreshToken(ctx, refreshToken)
	if err != nil {
		s.logger.Error("failed to refresh token without organization", "error", err)
		return "", err
	}

	if err := s.saveTokens(resp.AccessToken, resp.RefreshToken); err != nil {
		s.logger.Error("failed to save refreshed tokens", "error", err)
		return "", err
	}

	s.logger.Info("token refreshed without organization scope")
	return resp.AccessToken, nil
}

// RefreshTokenWithOrganization refreshes the access token scoped to an organization.
// This is used after creating/selecting an organization to get a token with the org_id claim.
// Returns the new access token so callers can update their API clients.
func (s *Service) RefreshTokenWithOrganization(ctx context.Context, workosOrgID domain.WorkosOrganizationID) (string, error) {
	s.logger.Debug("refreshing token with organization", "workos_org_id", workosOrgID)

	refreshToken, err := s.storage.Get("refresh_token")
	if err != nil {
		s.logger.Error("failed to get refresh token", "error", err)
		return "", err
	}
	if refreshToken == "" {
		return "", errors.New("no refresh token found")
	}

	resp, err := s.provider.RefreshTokenWithOrganization(ctx, refreshToken, workosOrgID)
	if err != nil {
		s.logger.Error("failed to refresh token with organization", "error", err)
		return "", err
	}

	if err := s.saveTokens(resp.AccessToken, resp.RefreshToken); err != nil {
		s.logger.Error("failed to save refreshed tokens", "error", err)
		return "", err
	}

	s.logger.Info("token refreshed with organization scope", "workos_org_id", workosOrgID)
	return resp.AccessToken, nil
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
