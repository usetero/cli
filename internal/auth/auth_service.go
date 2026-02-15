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
	GetUserID(ctx context.Context) (string, error)
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
	scope    log.Scope
}

// Ensure Service implements Auth.
var _ Auth = (*Service)(nil)

// NewService creates a new authentication service.
func NewService(provider OAuthProvider, storage SecureStorage, scope log.Scope) *Service {
	scope = scope.Child("auth")
	return &Service{
		provider: provider,
		storage:  storage,
		scope:    scope,
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
	s.scope.Debug("starting device authorization flow")

	resp, err := s.provider.AuthorizeDevice(ctx)
	if err != nil {
		s.scope.Error("failed to start device authorization", "error", err)
		return nil, err
	}

	s.scope.Debug("device authorization started",
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
	s.scope.Debug("starting authentication polling", "interval", interval)

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
				s.scope.Info("authentication successful", "user_id", resp.User.ID, "email", resp.User.Email)

				if err := s.saveTokens(resp.AccessToken, resp.RefreshToken); err != nil {
					s.scope.Error("failed to save tokens", "error", err)
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
				s.scope.Debug("authorization pending, continuing to poll")
				continue

			case errors.As(err, &slowDownErr):
				// Increase polling interval
				currentInterval = currentInterval * 2
				ticker.Reset(currentInterval)
				s.scope.Debug("slowing down polling", "new_interval", currentInterval)
				continue

			case errors.As(err, &expiredErr):
				s.scope.Error("device code expired")
				return nil, errors.New("device code expired - press 'r' to restart")

			case errors.As(err, &deniedErr):
				s.scope.Info("user denied authorization")
				return nil, errors.New("user denied authorization")

			default:
				// Unknown error
				s.scope.Error("authentication polling failed", "error", err)
				return nil, err
			}
		}
	}
}

// IsAuthenticated checks if the user has valid stored credentials.
func (s *Service) IsAuthenticated() bool {
	accessToken, err := s.storage.Get("access_token")
	if err != nil {
		s.scope.Error("failed to check authentication", "error", err)
		return false
	}
	return accessToken != ""
}

// GetUserID returns the WorkOS user ID from the current access token.
func (s *Service) GetUserID(ctx context.Context) (string, error) {
	token, err := s.GetAccessToken(ctx)
	if err != nil {
		return "", err
	}
	claims, err := ParseToken(token)
	if err != nil {
		return "", err
	}
	if claims.Sub == "" {
		return "", errors.New("token has no user ID")
	}
	return claims.Sub, nil
}

// GetAccessToken retrieves the stored access token, refreshing if expired.
func (s *Service) GetAccessToken(ctx context.Context) (string, error) {
	accessToken, err := s.storage.Get("access_token")
	if err != nil {
		s.scope.Error("failed to get access token", "error", err)
		return "", err
	}
	if accessToken == "" {
		return "", errors.New("no access token found")
	}

	// Check if token is expired
	claims, err := ParseToken(accessToken)
	if err != nil || claims.IsExpired() {
		s.scope.Debug("access token expired, refreshing")
		refreshToken, err := s.storage.Get("refresh_token")
		if err != nil {
			s.scope.Error("failed to get refresh token", "error", err)
			return "", err
		}
		if refreshToken == "" {
			return "", errors.New("no refresh token found")
		}

		resp, err := s.provider.RefreshToken(ctx, refreshToken)
		if err != nil {
			s.scope.Error("failed to refresh token", "error", err)
			return "", err
		}

		if err := s.saveTokens(resp.AccessToken, resp.RefreshToken); err != nil {
			s.scope.Error("failed to save refreshed tokens", "error", err)
			return "", err
		}

		s.scope.Debug("token refreshed successfully")
		return resp.AccessToken, nil
	}

	return accessToken, nil
}

// ForceRefreshAccessToken refreshes the access token unconditionally, bypassing
// the local expiration check. This is needed when a server rejects a token that
// the client still considers valid (e.g. due to clock skew).
func (s *Service) ForceRefreshAccessToken(ctx context.Context) (string, error) {
	s.scope.Debug("force-refreshing access token")

	refreshToken, err := s.storage.Get("refresh_token")
	if err != nil {
		s.scope.Error("failed to get refresh token", "error", err)
		return "", err
	}
	if refreshToken == "" {
		return "", errors.New("no refresh token found")
	}

	resp, err := s.provider.RefreshToken(ctx, refreshToken)
	if err != nil {
		s.scope.Error("failed to refresh token", "error", err)
		return "", err
	}

	if err := s.saveTokens(resp.AccessToken, resp.RefreshToken); err != nil {
		s.scope.Error("failed to save refreshed tokens", "error", err)
		return "", err
	}

	s.scope.Debug("token force-refreshed successfully")
	return resp.AccessToken, nil
}

// ClearTokens removes all stored authentication tokens.
func (s *Service) ClearTokens() error {
	s.scope.Info("clearing authentication tokens")
	if err := s.storage.Delete("access_token"); err != nil {
		s.scope.Error("failed to delete access token", "error", err)
		return err
	}
	if err := s.storage.Delete("refresh_token"); err != nil {
		s.scope.Error("failed to delete refresh token", "error", err)
		return err
	}
	return nil
}

// RefreshTokenWithoutOrganization refreshes the access token without any organization scope.
// This is used for bootstrap flows where the user needs a user-scoped token to create an org.
// Returns the new access token so callers can update their API clients.
func (s *Service) RefreshTokenWithoutOrganization(ctx context.Context) (string, error) {
	s.scope.Debug("refreshing token without organization scope")

	refreshToken, err := s.storage.Get("refresh_token")
	if err != nil {
		s.scope.Error("failed to get refresh token", "error", err)
		return "", err
	}
	if refreshToken == "" {
		return "", errors.New("no refresh token found")
	}

	resp, err := s.provider.RefreshToken(ctx, refreshToken)
	if err != nil {
		s.scope.Error("failed to refresh token without organization", "error", err)
		return "", err
	}

	if err := s.saveTokens(resp.AccessToken, resp.RefreshToken); err != nil {
		s.scope.Error("failed to save refreshed tokens", "error", err)
		return "", err
	}

	s.scope.Info("token refreshed without organization scope")
	return resp.AccessToken, nil
}

// RefreshTokenWithOrganization refreshes the access token scoped to an organization.
// This is used after creating/selecting an organization to get a token with the org_id claim.
// Returns the new access token so callers can update their API clients.
func (s *Service) RefreshTokenWithOrganization(ctx context.Context, workosOrgID domain.WorkosOrganizationID) (string, error) {
	s.scope.Debug("refreshing token with organization", "workos_org_id", workosOrgID)

	refreshToken, err := s.storage.Get("refresh_token")
	if err != nil {
		s.scope.Error("failed to get refresh token", "error", err)
		return "", err
	}
	if refreshToken == "" {
		return "", errors.New("no refresh token found")
	}

	resp, err := s.provider.RefreshTokenWithOrganization(ctx, refreshToken, workosOrgID)
	if err != nil {
		s.scope.Error("failed to refresh token with organization", "error", err)
		return "", err
	}

	if err := s.saveTokens(resp.AccessToken, resp.RefreshToken); err != nil {
		s.scope.Error("failed to save refreshed tokens", "error", err)
		return "", err
	}

	s.scope.Info("token refreshed with organization scope", "workos_org_id", workosOrgID)
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
