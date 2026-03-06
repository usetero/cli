package identity

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/usetero/cli/internal/infrastructure/auth/workos"
	"github.com/usetero/cli/internal/infrastructure/auth/workos/workostest"
)

func TestWorkOSProvider_Mapping(t *testing.T) {
	provider := NewWorkOSProvider(&workostest.Client{
		StartDeviceAuthorizationFn: func(context.Context) (workos.DeviceAuthorization, error) {
			return workos.DeviceAuthorization{
				UserCode:                "USER",
				VerificationURI:         "https://verify",
				VerificationURIComplete: "https://verify?code=USER",
				DeviceCode:              "DEVICE",
				ExpiresIn:               10 * time.Minute,
				Interval:                5 * time.Second,
			}, nil
		},
		PollDeviceAuthorizationFn: func(context.Context, string) (workos.DeviceAuthenticationResult, error) {
			return workos.DeviceAuthenticationResult{
				AccessToken:  "a1",
				RefreshToken: "r1",
				User: workos.AuthenticatedUser{
					ID:            "u_1",
					Email:         "u@example.com",
					EmailVerified: true,
					FirstName:     "U",
					LastName:      "Ser",
				},
			}, nil
		},
		RefreshTokenFn: func(context.Context, string, string) (workos.RefreshResult, error) {
			return workos.RefreshResult{AccessToken: "a2", RefreshToken: "r2"}, nil
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

func TestWorkOSProvider_ErrorMapping(t *testing.T) {
	provider := NewWorkOSProvider(&workostest.Client{
		PollDeviceAuthorizationFn: func(context.Context, string) (workos.DeviceAuthenticationResult, error) {
			return workos.DeviceAuthenticationResult{}, workos.ErrAuthorizationPending
		},
	})
	if _, _, err := provider.PollAuthentication(context.Background(), "DEVICE"); !errors.Is(err, ErrAuthorizationPending) {
		t.Fatalf("expected authorization pending mapping, got %v", err)
	}
}
