package workos_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/usetero/cli/internal/domains/identity"
	infraworkos "github.com/usetero/cli/internal/infrastructure/auth/workos"
	"github.com/usetero/cli/internal/infrastructure/auth/workos/workostest"
)

func TestProviderMapping(t *testing.T) {
	provider := infraworkos.NewProvider(&workostest.Client{
		StartDeviceAuthorizationFn: func(context.Context) (infraworkos.DeviceAuthorization, error) {
			return infraworkos.DeviceAuthorization{
				UserCode:                "USER",
				VerificationURI:         "https://verify",
				VerificationURIComplete: "https://verify?code=USER",
				DeviceCode:              "DEVICE",
				ExpiresIn:               10 * time.Minute,
				Interval:                5 * time.Second,
			}, nil
		},
		PollDeviceAuthorizationFn: func(context.Context, string) (infraworkos.DeviceAuthenticationResult, error) {
			return infraworkos.DeviceAuthenticationResult{
				AccessToken:  "a1",
				RefreshToken: "r1",
				User: infraworkos.AuthenticatedUser{
					ID:            "u_1",
					Email:         "u@example.com",
					EmailVerified: true,
					FirstName:     "U",
					LastName:      "Ser",
				},
			}, nil
		},
		RefreshTokenFn: func(context.Context, string, string) (infraworkos.RefreshResult, error) {
			return infraworkos.RefreshResult{AccessToken: "a2", RefreshToken: "r2"}, nil
		},
	})

	flow, err := provider.StartDeviceAuth(context.Background())
	if err != nil || flow.DeviceCode != "DEVICE" || flow.UserCode != "USER" {
		t.Fatalf("unexpected start flow=%+v err=%v", flow, err)
	}

	tokens, user, err := provider.PollAuthentication(context.Background(), "DEVICE")
	if err != nil {
		t.Fatalf("poll auth: %v", err)
	}
	if tokens.AccessToken != "a1" || tokens.RefreshToken != "r1" || user.ID != "u_1" {
		t.Fatalf("unexpected poll mapping tokens=%+v user=%+v", tokens, user)
	}

	refreshed, err := provider.Refresh(context.Background(), "r1", "org_1")
	if err != nil || refreshed.AccessToken != "a2" || refreshed.RefreshToken != "r2" {
		t.Fatalf("unexpected refresh mapping tokens=%+v err=%v", refreshed, err)
	}
}

func TestProviderErrorMapping(t *testing.T) {
	provider := infraworkos.NewProvider(&workostest.Client{
		PollDeviceAuthorizationFn: func(context.Context, string) (infraworkos.DeviceAuthenticationResult, error) {
			return infraworkos.DeviceAuthenticationResult{}, infraworkos.ErrAuthorizationPending
		},
	})
	if _, _, err := provider.PollAuthentication(context.Background(), "DEVICE"); !errors.Is(err, identity.ErrAuthorizationPending) {
		t.Fatalf("expected authorization pending mapping, got %v", err)
	}
}

func TestProviderRefreshInvalidGrantMapsToSessionExpired(t *testing.T) {
	provider := infraworkos.NewProvider(&workostest.Client{
		RefreshTokenFn: func(context.Context, string, string) (infraworkos.RefreshResult, error) {
			return infraworkos.RefreshResult{}, &infraworkos.OAuthError{
				Code:        "invalid_grant",
				Description: "Session has already ended.",
			}
		},
	})

	_, err := provider.Refresh(context.Background(), "r1", "")
	if !errors.Is(err, identity.ErrSessionExpired) {
		t.Fatalf("expected session expired mapping, got %v", err)
	}
}
