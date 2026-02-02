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
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
	"github.com/usetero/cli/internal/workos"
	"github.com/usetero/cli/pkg/client"
)

// Integration tests run against real PowerSync service.
//
// Prerequisites:
//  1. task auth:login
//  2. task run (complete onboarding to set default account)
//  3. task test:integration

func TestIntegration_Sync(t *testing.T) {
	cliConfig := config.LoadCLIConfig()
	logger := logtest.New(t)

	// Derive namespace from API endpoint (empty for production, host for dev)
	namespace := cliConfig.Namespace()
	if namespace == "" {
		namespace = "api.usetero.com" // production keyring namespace
	}

	t.Logf("Namespace: %s", namespace)
	t.Logf("API Endpoint: %s", cliConfig.APIEndpoint)
	t.Logf("PowerSync Endpoint: %s", cliConfig.PowerSyncEndpoint)

	// Get valid access token (refreshes if expired)
	storage := keyring.New(namespace)
	oauthProvider := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.ChatEndpoint, cliConfig.PowerSyncEndpoint)
	authSvc := auth.NewService(oauthProvider, storage, logger)

	token, err := authSvc.GetAccessToken(context.Background())
	if err != nil {
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

	// Validate account exists via API
	refreshFunc := func() (string, error) {
		return authSvc.GetAccessToken(context.Background())
	}
	apiClient := client.New(cliConfig.APIEndpoint, token, refreshFunc)
	apiClient.SetAccountID(accountID)
	accountSvc := api.NewAccountService(apiClient, logger)

	account, err := accountSvc.Get(context.Background(), accountID)
	if err != nil {
		t.Fatalf("Failed to fetch account: %v", err)
	}
	if account == nil {
		t.Fatalf("Account %s not found", accountID)
	}

	t.Run("connects and syncs data", func(t *testing.T) {
		t.Logf("Endpoint: %s", cliConfig.PowerSyncEndpoint)
		t.Logf("Account: %s (%s)", account.Name, account.ID)

		db := sqlitetest.OpenTest(t)
		sync := powersync.NewSync(cliConfig.PowerSyncEndpoint, authSvc, logtest.New(t))

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Wait for first sync to complete
		firstSyncDone := make(chan struct{})
		err := sync.Start(ctx, db, accountID, func() {
			close(firstSyncDone)
		})
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer sync.Stop()

		// Wait for first sync with status logging
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		var lastStatus powersync.Status
		for {
			select {
			case <-firstSyncDone:
				t.Logf("First sync completed")
				goto done
			case <-ticker.C:
				status := sync.Status()
				if status != lastStatus {
					t.Logf("Status: %s", status)
					if status == powersync.StatusError {
						t.Fatalf("Sync error: %v", sync.LastError())
					}
					lastStatus = status
				}
			case <-ctx.Done():
				t.Fatalf("Timeout waiting for first sync, last status: %s, error: %v", sync.Status(), sync.LastError())
			}
		}
	done:

		// Verify data was synced
		buckets, _ := countBuckets(db)
		services, _ := countServices(db)
		t.Logf("Synced %d services, %d buckets", services, buckets)

		if buckets == 0 {
			t.Fatal("No buckets synced - data not received from PowerSync")
		}
		if services == 0 {
			t.Fatal("No services synced - data not written to SQLite")
		}
	})
}

func countServices(db sqlite.Database) (int64, error) {
	var count int64
	err := db.DB().QueryRow(context.Background(), "SELECT COUNT(*) FROM services").Scan(&count)
	return count, err
}

func countBuckets(db sqlite.Database) (int64, error) {
	var count int64
	err := db.DB().QueryRow(context.Background(), "SELECT COUNT(*) FROM ps_buckets").Scan(&count)
	return count, err
}
