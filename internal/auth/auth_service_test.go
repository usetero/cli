package auth_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/auth/authtest"
	"github.com/usetero/cli/internal/log/logtest"
)

func TestService_GetAccessToken(t *testing.T) {
	t.Run("returns stored token when not expired", func(t *testing.T) {
		validToken := makeTestToken(time.Now().Add(10 * time.Minute))

		storage := &authtest.MockSecureStorage{
			GetFunc: func(key string) (string, error) {
				if key == "access_token" {
					return validToken, nil
				}
				return "", nil
			},
		}

		svc := auth.NewService(&authtest.MockOAuthProvider{}, storage, logtest.New(t))

		token, err := svc.GetAccessToken(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token != validToken {
			t.Errorf("got %q, want %q", token, validToken)
		}
	})

	t.Run("refreshes token when expired", func(t *testing.T) {
		expiredToken := makeTestToken(time.Now().Add(-10 * time.Minute))
		newToken := makeTestToken(time.Now().Add(10 * time.Minute))

		storage := &authtest.MockSecureStorage{
			GetFunc: func(key string) (string, error) {
				switch key {
				case "access_token":
					return expiredToken, nil
				case "refresh_token":
					return "refresh_token_value", nil
				}
				return "", nil
			},
		}

		provider := &authtest.MockOAuthProvider{
			RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*auth.RefreshResponse, error) {
				if refreshToken != "refresh_token_value" {
					t.Errorf("unexpected refresh token: %s", refreshToken)
				}
				return &auth.RefreshResponse{
					AccessToken:  newToken,
					RefreshToken: "new_refresh_token",
				}, nil
			},
		}

		svc := auth.NewService(provider, storage, logtest.New(t))

		token, err := svc.GetAccessToken(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token != newToken {
			t.Errorf("got %q, want %q", token, newToken)
		}
	})

	t.Run("returns error when refresh fails", func(t *testing.T) {
		expiredToken := makeTestToken(time.Now().Add(-10 * time.Minute))

		storage := &authtest.MockSecureStorage{
			GetFunc: func(key string) (string, error) {
				switch key {
				case "access_token":
					return expiredToken, nil
				case "refresh_token":
					return "refresh_token_value", nil
				}
				return "", nil
			},
		}

		provider := &authtest.MockOAuthProvider{
			RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*auth.RefreshResponse, error) {
				return nil, errors.New("refresh failed")
			},
		}

		svc := auth.NewService(provider, storage, logtest.New(t))

		_, err := svc.GetAccessToken(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestService_RefreshTokenWithoutOrganization(t *testing.T) {
	t.Run("refreshes token without org scope", func(t *testing.T) {
		newToken := makeTestToken(time.Now().Add(10 * time.Minute))
		var savedAccessToken, savedRefreshToken string

		storage := &authtest.MockSecureStorage{
			GetFunc: func(key string) (string, error) {
				if key == "refresh_token" {
					return "stored_refresh_token", nil
				}
				return "", nil
			},
			SetFunc: func(key, value string) error {
				switch key {
				case "access_token":
					savedAccessToken = value
				case "refresh_token":
					savedRefreshToken = value
				}
				return nil
			},
		}

		provider := &authtest.MockOAuthProvider{
			RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*auth.RefreshResponse, error) {
				if refreshToken != "stored_refresh_token" {
					t.Errorf("unexpected refresh token: %s", refreshToken)
				}
				return &auth.RefreshResponse{
					AccessToken:  newToken,
					RefreshToken: "new_refresh_token",
				}, nil
			},
		}

		svc := auth.NewService(provider, storage, logtest.New(t))

		token, err := svc.RefreshTokenWithoutOrganization(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token != newToken {
			t.Errorf("got %q, want %q", token, newToken)
		}
		if savedAccessToken != newToken {
			t.Errorf("access token not saved: got %q, want %q", savedAccessToken, newToken)
		}
		if savedRefreshToken != "new_refresh_token" {
			t.Errorf("refresh token not saved: got %q, want %q", savedRefreshToken, "new_refresh_token")
		}
	})

	t.Run("returns error when no refresh token", func(t *testing.T) {
		storage := &authtest.MockSecureStorage{
			GetFunc: func(key string) (string, error) {
				return "", nil
			},
		}

		svc := auth.NewService(&authtest.MockOAuthProvider{}, storage, logtest.New(t))

		_, err := svc.RefreshTokenWithoutOrganization(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// makeTestToken creates a JWT with the given expiration time.
func makeTestToken(exp time.Time) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256"}`))
	payload, _ := json.Marshal(map[string]int64{"exp": exp.Unix()})
	payloadEnc := base64.RawURLEncoding.EncodeToString(payload)
	sig := base64.RawURLEncoding.EncodeToString([]byte("signature"))
	return header + "." + payloadEnc + "." + sig
}
