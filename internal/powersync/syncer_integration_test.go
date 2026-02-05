//go:build integration

package powersync_test

import (
	"context"
	"testing"
	"time"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/keyring"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/powersync/db/dbtest"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/workos"
)

// Integration tests run against real PowerSync service.
//
// Prerequisites:
//  1. task auth:login
//  2. task run (complete onboarding to set default account)
//  3. task test:integration

func TestIntegration_Syncer(t *testing.T) {
	cliConfig := config.LoadCLIConfig()
	logger := logtest.NewScope(t)

	// Derive namespace from API endpoint (empty for production, host for dev)
	namespace := cliConfig.Namespace()
	if namespace == "" {
		namespace = "api.usetero.com" // production keyring namespace
	}

	t.Logf("Namespace: %s", namespace)
	t.Logf("API Endpoint: %s", cliConfig.APIEndpoint)
	t.Logf("PowerSync Endpoint: %s", cliConfig.PowerSyncEndpoint)

	// Get auth service
	storage := keyring.New(namespace)
	oauthProvider := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.ChatEndpoint, cliConfig.PowerSyncEndpoint)
	authSvc := auth.NewService(oauthProvider, storage, logger)

	// Verify we have valid credentials
	if _, err := authSvc.GetAccessToken(context.Background()); err != nil {
		t.Fatalf("Failed to get access token: %v (run: task auth:login)", err)
	}

	// Get account ID from preferences
	cfg, err := config.Load(namespace)
	if err != nil {
		t.Fatalf("Config not found: %v (run: task run)", err)
	}
	prefs := preferences.NewService(cfg, logger)
	accountID := prefs.GetDefaultAccountID()
	if accountID == "" {
		t.Fatalf("No default account (run: task run)")
	}

	t.Logf("Account ID from config: %s", accountID)

	// Create API services
	services := api.NewServices(cliConfig.APIEndpoint+"/graphql", authSvc, logger)
	services.SetAccountID(domain.AccountID(accountID))

	// Validate account exists via API
	account, err := services.Accounts.Get(context.Background(), accountID)
	if err != nil {
		t.Fatalf("Failed to fetch account: %v", err)
	}
	if account == nil {
		t.Fatalf("Account %s not found", accountID)
	}

	t.Run("connects and syncs data", func(t *testing.T) {
		t.Logf("Endpoint: %s", cliConfig.PowerSyncEndpoint)
		t.Logf("Account: %s (%s)", account.Name, account.ID)

		db := dbtest.OpenTestDB(t)
		syncer := powersync.NewSyncer(cliConfig.PowerSyncEndpoint, authSvc, logtest.NewScope(t))

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Wait for first sync to complete
		firstSyncDone := make(chan struct{})
		err := syncer.Start(ctx, db, accountID.String(), func() {
			close(firstSyncDone)
		})
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer syncer.Stop()

		// Wait for first sync with status logging
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		var lastPhase powersync.Phase
		for {
			select {
			case <-firstSyncDone:
				t.Logf("First sync completed")
				goto done
			case <-ticker.C:
				state := syncer.State()
				if state.Phase != lastPhase {
					t.Logf("Phase: %s", state.Phase)
					if state.Phase == powersync.PhaseError {
						t.Fatalf("Sync error: %v", state.Error)
					}
					lastPhase = state.Phase
				}
			case <-ctx.Done():
				t.Fatalf("Timeout waiting for first sync, last phase: %s, error: %v", syncer.State().Phase, syncer.State().Error)
			}
		}
	done:

		// Verify data was synced
		buckets, _ := countBuckets(db)
		t.Logf("Synced %d buckets", buckets)

		if buckets == 0 {
			t.Fatal("No buckets synced - data not received from PowerSync")
		}

		// Verify IsReady is true after sync
		if !syncer.IsReady() {
			t.Error("IsReady() should be true after first sync")
		}
	})
}

func countBuckets(db sqlite.DB) (int64, error) {
	var count int64
	err := db.QueryRow(context.Background(), "SELECT COUNT(*) FROM ps_buckets").Scan(&count)
	return count, err
}
