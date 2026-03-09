//go:build integration_live

package main

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/usetero/cli/cmd/tero/terotest"
)

func TestIntegrationLive_Tero_OnboardingHappyPath(t *testing.T) {
	if env := strings.TrimSpace(os.Getenv("TERO_ENV")); env != "" && env != "local" {
		t.Skipf("local onboarding E2E requires TERO_ENV=local, got %q", env)
	}

	homeDir := t.TempDir()
	run := terotest.RequireLocalRun(t, homeDir)
	binary := terotest.Build(t)
	app := terotest.Start(t, binary, terotest.Options{
		HomeDir: homeDir,
		Env:     run.AppEnv,
	})
	defer app.Stop()

	app.WaitFor("Select your role:", 15*time.Second)
	app.Press(terotest.KeyEnter)

	app.WaitFor("Create your organization:", 30*time.Second)
	app.Type(run.OrganizationName)
	app.Press(terotest.KeyEnter)

	app.WaitFor("Select your Datadog site:", 90*time.Second)
	orgID, err := run.FindOrganizationID(context.Background(), run.OrganizationName)
	if err != nil {
		t.Fatalf("resolve created organization: %v", err)
	}
	defer func() {
		if err := run.DeleteOrganization(context.Background(), orgID); err != nil {
			t.Fatalf("cleanup organization %s: %v", orgID, err)
		}
	}()
	app.Press(terotest.KeyEnter)

	app.WaitFor("Enter Datadog API key:", 90*time.Second)
	app.Type(run.DatadogAPIKey)
	app.Press(terotest.KeyEnter)

	app.WaitFor("Finish Datadog setup:", 90*time.Second)
	app.Type(run.DatadogName)
	app.Press(terotest.KeyTab)
	app.Type(run.DatadogAppKey)
	app.Press(terotest.KeyEnter)

	app.WaitFor("Waiting for Datadog discovery...", 2*time.Minute)
	app.Press(terotest.KeyEnter)

	app.WaitFor("Syncing your account data...", 2*time.Minute)
	app.WaitForAny(3*time.Minute, "PowerSync is ready.", "Welcome to Tero")
	app.WaitFor("Welcome to Tero", 3*time.Minute)

	snapshot := app.Snapshot()
	if strings.Contains(snapshot, "Failed to load onboarding state.") {
		t.Fatalf("unexpected onboarding error:\n%s", snapshot)
	}
}
