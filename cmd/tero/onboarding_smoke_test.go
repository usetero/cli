//go:build integration

package main

import (
	"strings"
	"testing"
	"time"

	"github.com/usetero/cli/cmd/tero/terotest"
	domainprefs "github.com/usetero/cli/internal/domains/preferences"
)

func TestIntegration_Tero_BootsToOrganizationCreation(t *testing.T) {
	services := terotest.StartFakeServices(t)
	homeDir := t.TempDir()
	secretStorePath := terotest.SecretStorePath(homeDir, "local")
	terotest.SeedTokens(t, homeDir, "local", services.AccessToken, services.RefreshToken)

	binary := terotest.Build(t)
	app := terotest.Start(t, binary, terotest.Options{
		HomeDir: homeDir,
		Env: map[string]string{
			"TERO_ENV":               "local",
			"TERO_SECRET_STORE":      "file",
			"TERO_SECRET_STORE_PATH": secretStorePath,
			"TERO_API_ORIGIN":        services.APIOrigin,
			"TERO_CHAT_ORIGIN":       services.ChatOrigin,
			"TERO_POWERSYNC_ORIGIN":  services.PowerSyncOrigin,
		},
	})
	defer app.Stop()

	app.WaitFor("Create your organization.", 5*time.Second)

	if strings.Contains(app.Snapshot(), "Failed to load onboarding state.") {
		t.Fatalf("unexpected onboarding error:\n%s", app.Snapshot())
	}
}

func TestIntegration_Tero_IgnoresStaleOrganizationPreference(t *testing.T) {
	services := terotest.StartFakeServices(t,
		terotest.WithFakeOrganizations([]terotest.FakeOrganization{
			{ID: "org_1", Name: "Org 1"},
			{ID: "org_2", Name: "Org 2"},
		}),
	)
	homeDir := t.TempDir()
	secretStorePath := terotest.SecretStorePath(homeDir, "local")
	terotest.SeedTokens(t, homeDir, "local", services.AccessToken, services.RefreshToken)
	terotest.SeedPreferences(t, homeDir, "local", domainprefs.Snapshot{
		Organization: "org_stale",
	})

	binary := terotest.Build(t)
	app := terotest.Start(t, binary, terotest.Options{
		HomeDir: homeDir,
		Env: map[string]string{
			"TERO_ENV":               "local",
			"TERO_SECRET_STORE":      "file",
			"TERO_SECRET_STORE_PATH": secretStorePath,
			"TERO_API_ORIGIN":        services.APIOrigin,
			"TERO_CHAT_ORIGIN":       services.ChatOrigin,
			"TERO_POWERSYNC_ORIGIN":  services.PowerSyncOrigin,
		},
	})
	defer app.Stop()

	app.WaitFor("Choose your organization.", 5*time.Second)

	if strings.Contains(app.Snapshot(), "Create your organization.") {
		t.Fatalf("expected stale preference to route to selection, got:\n%s", app.Snapshot())
	}
}

func TestIntegration_Tero_ExistingDatadogSkipsSetup(t *testing.T) {
	services := terotest.StartFakeServices(t,
		terotest.WithFakeOrganizations([]terotest.FakeOrganization{
			{
				ID:   "org_1",
				Name: "Org 1",
				Accounts: []terotest.FakeAccount{
					{
						ID:   "acct_1",
						Name: "Account 1",
						Workspaces: []terotest.FakeWorkspace{
							{ID: "ws_1", Name: "Workspace 1"},
						},
						Datadog: &terotest.FakeDatadogAccount{
							ID:   "dd_1",
							Name: "Datadog Demo",
							Site: "US5",
							Status: terotest.FakeDatadogStatus{
								ReadyForUse:    false,
								EventCount:     100,
								AnalyzedCount:  42,
								ServiceCount:   3,
								ActiveServices: 2,
							},
						},
					},
				},
			},
		}),
	)
	homeDir := t.TempDir()
	secretStorePath := terotest.SecretStorePath(homeDir, "local")
	terotest.SeedTokens(t, homeDir, "local", services.AccessToken, services.RefreshToken)

	binary := terotest.Build(t)
	app := terotest.Start(t, binary, terotest.Options{
		HomeDir: homeDir,
		Env: map[string]string{
			"TERO_ENV":               "local",
			"TERO_SECRET_STORE":      "file",
			"TERO_SECRET_STORE_PATH": secretStorePath,
			"TERO_API_ORIGIN":        services.APIOrigin,
			"TERO_CHAT_ORIGIN":       services.ChatOrigin,
			"TERO_POWERSYNC_ORIGIN":  services.PowerSyncOrigin,
		},
	})
	defer app.Stop()

	app.WaitFor("We're discovering your Datadog events. This can take a few minutes.", 5*time.Second)

	snapshot := app.Snapshot()
	if strings.Contains(snapshot, "Choose your Datadog region.") {
		t.Fatalf("expected existing datadog account to skip setup, got:\n%s", snapshot)
	}
}

func TestIntegration_Tero_ExpiredAuthRoutesBackToSignIn(t *testing.T) {
	services := terotest.StartFakeServices(t)
	homeDir := t.TempDir()
	secretStorePath := terotest.SecretStorePath(homeDir, "local")
	terotest.SeedTokens(t, homeDir, "local", terotest.AccessTokenWithTTL(t, -1*time.Minute), "")

	binary := terotest.Build(t)
	app := terotest.Start(t, binary, terotest.Options{
		HomeDir: homeDir,
		Env: map[string]string{
			"TERO_ENV":               "local",
			"TERO_SECRET_STORE":      "file",
			"TERO_SECRET_STORE_PATH": secretStorePath,
			"TERO_API_ORIGIN":        services.APIOrigin,
			"TERO_CHAT_ORIGIN":       services.ChatOrigin,
			"TERO_POWERSYNC_ORIGIN":  services.PowerSyncOrigin,
		},
	})
	defer app.Stop()

	app.WaitFor("Your session ended. Sign in again.", 5*time.Second)
}
