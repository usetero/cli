package identity_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	identity "github.com/usetero/cli/internal/domains/identity"
	"github.com/usetero/cli/internal/domains/identity/identitytest"
)

func TestPollDeviceFlow_PendingThenSuccess(t *testing.T) {
	t.Parallel()

	calls := 0
	s := identity.NewService(identitytest.Provider{
		PollAuthenticationFn: func(context.Context, string) (identity.Tokens, identity.User, error) {
			calls++
			if calls == 1 {
				return identity.Tokens{}, identity.User{}, identity.ErrAuthorizationPending
			}
			return identity.Tokens{AccessToken: identity.AccessToken("a"), RefreshToken: identity.RefreshToken("r")}, identity.User{ID: "u_1"}, nil
		},
		StartDeviceAuthFn: func(context.Context) (identity.DeviceFlow, error) { return identity.DeviceFlow{}, nil },
		RefreshFn: func(context.Context, identity.RefreshToken, identity.ProviderOrgID) (identity.Tokens, error) {
			return identity.Tokens{}, nil
		},
	}, &identitytest.TokenStore{}, identity.NopLogger{})

	user, err := s.PollDeviceFlow(context.Background(), "dc", 1*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "u_1" {
		t.Fatalf("unexpected user id: %s", user.ID)
	}
	if calls < 2 {
		t.Fatalf("expected at least 2 polls, got %d", calls)
	}
}

func TestGetAccessToken_RefreshesExpired(t *testing.T) {
	t.Parallel()

	store := &identitytest.TokenStore{
		AccessTokenValue:  identity.AccessToken(makeJWT(time.Now().Add(-1 * time.Minute))),
		RefreshTokenValue: identity.RefreshToken("refresh_1"),
	}

	s := identity.NewService(identitytest.Provider{
		StartDeviceAuthFn: func(context.Context) (identity.DeviceFlow, error) { return identity.DeviceFlow{}, nil },
		PollAuthenticationFn: func(context.Context, string) (identity.Tokens, identity.User, error) {
			return identity.Tokens{}, identity.User{}, nil
		},
		RefreshFn: func(_ context.Context, refreshToken identity.RefreshToken, providerOrgID identity.ProviderOrgID) (identity.Tokens, error) {
			if refreshToken != identity.RefreshToken("refresh_1") {
				return identity.Tokens{}, fmt.Errorf("unexpected refresh token")
			}
			if providerOrgID != "" {
				return identity.Tokens{}, fmt.Errorf("unexpected org id")
			}
			return identity.Tokens{AccessToken: identity.AccessToken("new_access"), RefreshToken: identity.RefreshToken("new_refresh")}, nil
		},
	}, store, identity.NopLogger{})

	got, err := s.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "new_access" {
		t.Fatalf("unexpected access token: %s", got)
	}
	if store.RefreshTokenValue != identity.RefreshToken("new_refresh") {
		t.Fatalf("refresh token not updated")
	}
}

func TestRefreshForOrganization(t *testing.T) {
	t.Parallel()

	store := &identitytest.TokenStore{RefreshTokenValue: identity.RefreshToken("refresh_1")}

	s := identity.NewService(identitytest.Provider{
		StartDeviceAuthFn: func(context.Context) (identity.DeviceFlow, error) { return identity.DeviceFlow{}, nil },
		PollAuthenticationFn: func(context.Context, string) (identity.Tokens, identity.User, error) {
			return identity.Tokens{}, identity.User{}, nil
		},
		RefreshFn: func(_ context.Context, refreshToken identity.RefreshToken, providerOrgID identity.ProviderOrgID) (identity.Tokens, error) {
			if refreshToken != identity.RefreshToken("refresh_1") || providerOrgID != identity.ProviderOrgID("org_1") {
				return identity.Tokens{}, errors.New("bad refresh call")
			}
			return identity.Tokens{AccessToken: identity.AccessToken("org_access"), RefreshToken: identity.RefreshToken("org_refresh")}, nil
		},
	}, store, identity.NopLogger{})

	got, err := s.RefreshForOrganization(context.Background(), identity.ProviderOrgID("org_1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "org_access" {
		t.Fatalf("unexpected token: %s", got)
	}
}

func TestSignOut(t *testing.T) {
	t.Parallel()

	store := &identitytest.TokenStore{
		AccessTokenValue:  identity.AccessToken("a"),
		RefreshTokenValue: identity.RefreshToken("r"),
	}

	s := identity.NewService(identitytest.Provider{
		StartDeviceAuthFn: func(context.Context) (identity.DeviceFlow, error) { return identity.DeviceFlow{}, nil },
		PollAuthenticationFn: func(context.Context, string) (identity.Tokens, identity.User, error) {
			return identity.Tokens{}, identity.User{}, nil
		},
		RefreshFn: func(context.Context, identity.RefreshToken, identity.ProviderOrgID) (identity.Tokens, error) {
			return identity.Tokens{}, nil
		},
	}, store, identity.NopLogger{})

	if err := s.SignOut(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.AccessTokenValue != "" || store.RefreshTokenValue != "" {
		t.Fatalf("tokens were not cleared")
	}
}

func TestGetAccessToken_ReturnsCachedTokenWhenNotExpired(t *testing.T) {
	t.Parallel()

	store := &identitytest.TokenStore{
		AccessTokenValue:  identity.AccessToken(makeJWT(time.Now().Add(10 * time.Minute))),
		RefreshTokenValue: identity.RefreshToken("refresh_1"),
	}

	refreshCalled := false
	s := identity.NewService(identitytest.Provider{
		StartDeviceAuthFn: func(context.Context) (identity.DeviceFlow, error) { return identity.DeviceFlow{}, nil },
		PollAuthenticationFn: func(context.Context, string) (identity.Tokens, identity.User, error) {
			return identity.Tokens{}, identity.User{}, nil
		},
		RefreshFn: func(context.Context, identity.RefreshToken, identity.ProviderOrgID) (identity.Tokens, error) {
			refreshCalled = true
			return identity.Tokens{}, nil
		},
	}, store, identity.NopLogger{})

	got, err := s.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("get access token: %v", err)
	}
	if got != string(store.AccessTokenValue) {
		t.Fatalf("expected cached access token, got %q", got)
	}
	if refreshCalled {
		t.Fatalf("refresh should not be called when token is still valid")
	}
}

func TestGetAccessToken_ExpiredTokenWithoutRefreshReturnsNotAuthenticated(t *testing.T) {
	t.Parallel()

	s := identity.NewService(identitytest.Provider{
		StartDeviceAuthFn: func(context.Context) (identity.DeviceFlow, error) { return identity.DeviceFlow{}, nil },
		PollAuthenticationFn: func(context.Context, string) (identity.Tokens, identity.User, error) {
			return identity.Tokens{}, identity.User{}, nil
		},
		RefreshFn: func(context.Context, identity.RefreshToken, identity.ProviderOrgID) (identity.Tokens, error) {
			return identity.Tokens{}, nil
		},
	}, &identitytest.TokenStore{
		AccessTokenValue:  identity.AccessToken(makeJWT(time.Now().Add(-1 * time.Minute))),
		RefreshTokenValue: "",
	}, identity.NopLogger{})

	_, err := s.GetAccessToken(context.Background())
	if !errors.Is(err, identity.ErrNotAuthenticated) {
		t.Fatalf("expected ErrNotAuthenticated, got %v", err)
	}
}

func makeJWT(exp time.Time) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload, _ := json.Marshal(map[string]any{"exp": exp.Unix()})
	claims := base64.RawURLEncoding.EncodeToString(payload)
	return header + "." + claims + ".sig"
}
