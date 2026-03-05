package identity

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Service owns authentication lifecycle and token storage.
type Service struct {
	provider Provider
	store    TokenStore
	log      Logger
}

// NewService constructs an identity service.
func NewService(provider Provider, store TokenStore, log Logger) *Service {
	if log == nil {
		log = NopLogger{}
	}
	return &Service{
		provider: provider,
		store:    store,
		log:      log,
	}
}

// StartDeviceFlow starts the WorkOS device authorization flow.
func (s *Service) StartDeviceFlow(ctx context.Context) (DeviceFlow, error) {
	flow, err := s.provider.StartDeviceAuth(ctx)
	if err != nil {
		return DeviceFlow{}, err
	}
	return flow, nil
}

// PollDeviceFlow blocks until auth succeeds or fails.
func (s *Service) PollDeviceFlow(ctx context.Context, deviceCode string, interval time.Duration) (User, error) {
	if interval <= 0 {
		interval = 5 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	currentInterval := interval

	for {
		select {
		case <-ctx.Done():
			return User{}, ctx.Err()
		case <-ticker.C:
			tokens, user, err := s.provider.PollAuthentication(ctx, deviceCode)
			if err == nil {
				if err := s.saveTokens(tokens); err != nil {
					return User{}, err
				}
				s.log.Info("authentication successful", "user_id", user.ID, "email", user.Email)
				return user, nil
			}

			switch {
			case errors.Is(err, ErrAuthorizationPending):
				continue
			case errors.Is(err, ErrSlowDown):
				currentInterval = currentInterval * 2
				ticker.Reset(currentInterval)
				s.log.Debug("slowing auth polling", "interval", currentInterval.String())
				continue
			case errors.Is(err, ErrExpiredDeviceCode), errors.Is(err, ErrAccessDenied):
				return User{}, err
			default:
				return User{}, err
			}
		}
	}
}

// IsAuthenticated reports whether an access token exists.
func (s *Service) IsAuthenticated() bool {
	token, err := s.store.AccessToken()
	return err == nil && token != ""
}

// GetAccessToken returns a valid access token, refreshing if expired.
func (s *Service) GetAccessToken(ctx context.Context) (string, error) {
	accessToken, err := s.store.AccessToken()
	if err != nil {
		return "", err
	}
	if accessToken == "" {
		return "", ErrNotAuthenticated
	}

	claims, err := parseTokenClaims(string(accessToken))
	if err == nil && !claims.isExpired() {
		return string(accessToken), nil
	}

	refreshToken, err := s.store.RefreshToken()
	if err != nil {
		return "", err
	}
	if refreshToken == "" {
		return "", ErrNotAuthenticated
	}

	tokens, err := s.provider.Refresh(ctx, refreshToken, "")
	if err != nil {
		return "", err
	}
	if err := s.saveTokens(tokens); err != nil {
		return "", err
	}
	return string(tokens.AccessToken), nil
}

// RefreshForOrganization refreshes tokens scoped to the given provider organization.
func (s *Service) RefreshForOrganization(ctx context.Context, providerOrgID ProviderOrgID) (string, error) {
	refreshToken, err := s.store.RefreshToken()
	if err != nil {
		return "", err
	}
	if refreshToken == "" {
		return "", ErrNotAuthenticated
	}

	tokens, err := s.provider.Refresh(ctx, refreshToken, providerOrgID)
	if err != nil {
		return "", err
	}
	if err := s.saveTokens(tokens); err != nil {
		return "", err
	}
	return string(tokens.AccessToken), nil
}

// SignOut clears local tokens.
func (s *Service) SignOut() error {
	return s.store.ClearTokens()
}

func (s *Service) saveTokens(tokens Tokens) error {
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		return fmt.Errorf("tokens are incomplete")
	}
	return s.store.SetTokens(tokens)
}

type tokenClaims struct {
	Exp int64 `json:"exp"`
}

func (c tokenClaims) isExpired() bool {
	return time.Now().Unix() > c.Exp-30
}

func parseTokenClaims(token string) (tokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return tokenClaims{}, fmt.Errorf("invalid token format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return tokenClaims{}, err
	}
	var claims tokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return tokenClaims{}, err
	}
	return claims, nil
}
