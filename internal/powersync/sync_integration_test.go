//go:build integration

package powersync_test

import (
	"context"
	"os"
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
	powersyncEndpoint := os.Getenv("TERO_POWERSYNC_ENDPOINT")
	if powersyncEndpoint == "" {
		t.Fatalf("TERO_POWERSYNC_ENDPOINT not set")
	}

	apiEndpoint := os.Getenv("TERO_API_ENDPOINT")
	if apiEndpoint == "" {
		t.Fatalf("TERO_API_ENDPOINT not set")
	}

	workosClientID := os.Getenv("TERO_WORKOS_CLIENT_ID")
	if workosClientID == "" {
		t.Fatalf("TERO_WORKOS_CLIENT_ID not set")
	}

	// Audience URLs for JWT (base URL without path)
	apiAudience := "https://api.usetero.dev"

	namespace := "api.usetero.dev"
	logger := logtest.New(t)

	// Get valid access token (refreshes if expired)
	storage := keyring.New(namespace)
	oauthProvider := workos.NewClient(workosClientID, apiAudience, powersyncEndpoint)
	authSvc := auth.NewService(oauthProvider, storage, logger)

	token, err := authSvc.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("Failed to get access token: %v (run: task auth:login)", err)
	}

	// Get account ID from config
	cfg, err := config.Load(namespace)
	if err != nil {
		t.Fatalf("Config not found: %v (run: task run)", err)
	}
	accountID := cfg.Get("default_account_id")
	if accountID == "" {
		t.Fatalf("No default account (run: task run)")
	}

	// Validate account exists via API
	refreshFunc := func() (string, error) {
		return authSvc.GetAccessToken(context.Background())
	}
	apiClient := client.New(apiEndpoint, token, refreshFunc)
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
		t.Logf("Endpoint: %s", powersyncEndpoint)
		t.Logf("Account: %s (%s)", account.Name, account.ID)

		db := sqlitetest.OpenTest(t)
		psConfig := &powersync.Config{Endpoint: powersyncEndpoint}
		sync := powersync.NewSync(psConfig, nil, logtest.New(t))

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := sync.Start(ctx, db, accountID, token)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer sync.Stop()

		// Wait for sync to reach syncing state
		deadline := time.Now().Add(10 * time.Second)
		var lastStatus powersync.Status
		var reachedSyncing bool
		for time.Now().Before(deadline) {
			status := sync.Status()
			if status != lastStatus {
				t.Logf("Status: %s", status)
				if status == powersync.StatusError {
					t.Logf("Error: %v", sync.LastError())
				}
				lastStatus = status
			}

			if status == powersync.StatusSyncing {
				reachedSyncing = true
			}

			if status == powersync.StatusError {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}

		if !reachedSyncing {
			t.Fatalf("Never reached syncing state, last status: %s, error: %v", sync.Status(), sync.LastError())
		}

		if sync.Status() == powersync.StatusError {
			t.Fatalf("Sync failed: %v", sync.LastError())
		}

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
	err := db.QueryRow("SELECT COUNT(*) FROM services").Scan(&count)
	return count, err
}

func countBuckets(db sqlite.Database) (int64, error) {
	var count int64
	err := db.QueryRow("SELECT COUNT(*) FROM ps_buckets").Scan(&count)
	return count, err
}
