//go:build integration

package main

import (
	"strings"
	"testing"
	"time"

	"github.com/usetero/cli/cmd/tero/terotest"
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
