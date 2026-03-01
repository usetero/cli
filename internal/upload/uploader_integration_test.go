//go:build integration

package upload_test

import (
	"context"
	"testing"
	"time"

	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/boundary/chat"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/keyring"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
	"github.com/usetero/cli/internal/upload"
	"github.com/usetero/cli/internal/workos"
)

// Integration tests run against real services.
//
// Prerequisites:
//  1. task auth:login
//  2. task run (complete onboarding to set default account/workspace)
//  3. task test:integration

func TestIntegration_Upload(t *testing.T) {
	ctx := context.Background()
	logger := logtest.NewScope(t)

	// Load config
	cliConfig := config.LoadCLIConfig()
	env := cliConfig.Environment()

	t.Logf("Environment: %s", env)
	t.Logf("API Endpoint: %s", cliConfig.APIEndpoint)
	t.Logf("Chat Endpoint: %s", cliConfig.ChatEndpoint)
	t.Logf("PowerSync Endpoint: %s", cliConfig.PowerSyncEndpoint)

	// Get auth service
	storage := keyring.New(env)
	oauthProvider := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.ChatEndpoint, cliConfig.PowerSyncEndpoint)
	authSvc := auth.NewService(oauthProvider, storage, logger)

	// Verify we have valid credentials
	if _, err := authSvc.GetAccessToken(ctx); err != nil {
		t.Fatalf("Failed to get access token: %v (run: task auth:login)", err)
	}

	// Get account and workspace from preferences
	orgCfg, err := config.LoadOrgPreferences(env, config.ActiveOrgID(env))
	if err != nil {
		t.Fatalf("Org preferences not found: %v (run: task run)", err)
	}
	orgPrefs := preferences.NewOrgService(orgCfg, logger)

	accountID := orgPrefs.GetDefaultAccountID()
	if accountID == "" {
		t.Fatalf("No default account (run: task run)")
	}

	workspaceID := prefs.GetDefaultWorkspaceID()
	if workspaceID == "" {
		t.Fatalf("No default workspace (run: task run)")
	}

	t.Logf("Account ID: %s", accountID)
	t.Logf("Workspace ID: %s", workspaceID)

	// Create API services
	services := graphql.NewServiceSet(cliConfig.APIEndpoint+"/graphql", authSvc, logger)
	services.SetAccountID(domain.AccountID(accountID))

	// Create chat client
	chatClient := chat.NewClient(cliConfig.ChatEndpoint, authSvc, logger)
	chatClient.SetAccountID(accountID)

	// Create message service
	messagesSvc := graphql.NewMessageService(chatClient, logger)

	t.Run("mutation round-trip maintains healthy database", func(t *testing.T) {
		// Create fresh database with PowerSync
		db := sqlitetest.OpenBareDB(t)

		if err := powersync.ApplySchema(ctx, db); err != nil {
			t.Fatalf("ApplySchema() error = %v", err)
		}

		// Start PowerSync
		sync := powersync.NewSyncer(cliConfig.PowerSyncEndpoint, authSvc, logger)

		syncCtx, syncCancel := context.WithTimeout(ctx, 60*time.Second)
		defer syncCancel()

		firstSyncDone := make(chan struct{})
		err = sync.Start(syncCtx, db, accountID, func() {
			close(firstSyncDone)
		})
		if err != nil {
			t.Fatalf("Failed to start sync: %v", err)
		}
		defer sync.Stop()

		// Wait for initial sync
		t.Log("Waiting for initial sync...")
		select {
		case <-firstSyncDone:
			t.Log("Initial sync complete")
		case <-syncCtx.Done():
			t.Fatalf("Timeout waiting for initial sync: %v", sync.LastError())
		}

		// Verify database is healthy before mutation
		queue := powersync.NewCrudQueue(db)
		if err := queue.CheckHealth(ctx); err != nil {
			t.Fatalf("Database unhealthy before mutation: %v", err)
		}
		t.Log("Database healthy before mutation")

		// Start upload loop
		powersyncClient := powersync.NewClient(cliConfig.PowerSyncEndpoint)
		uploader := upload.New(
			db,
			powersyncClient,
			authSvc,
			upload.MutationDeps{
				Conversations: services.Conversations,
				Messages:      messagesSvc,
				Services:      services.Services,
				Policies:      services.Policies,
			},
			logger,
		)

		uploadCtx, uploadCancel := context.WithCancel(ctx)
		defer uploadCancel()

		uploadDone := make(chan error, 1)
		go func() {
			uploadDone <- uploader.Run(uploadCtx)
		}()

		// Create a conversation locally (this triggers CRUD entry)
		t.Log("Creating conversation...")
		convID, err := db.Conversations().Create(ctx, accountID, workspaceID)
		if err != nil {
			t.Fatalf("Failed to create conversation: %v", err)
		}
		t.Logf("Created conversation: %s", convID)

		// Create a message locally (this triggers CRUD entry)
		t.Log("Creating message...")
		msgID, err := db.Messages().CreateUserMessage(ctx, accountID, convID, "Hello from integration test")
		if err != nil {
			t.Fatalf("Failed to create message: %v", err)
		}
		t.Logf("Created message: %s", msgID)

		// Wait for CRUD queue to drain (upload processed entries)
		t.Log("Waiting for CRUD queue to drain...")
		drainTimeout := time.After(30 * time.Second)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-drainTimeout:
				// Check what's stuck
				hasPending, _ := queue.HasPendingUploads(ctx)
				t.Fatalf("Timeout waiting for CRUD queue to drain (hasPending=%v)", hasPending)
			case <-ticker.C:
				hasPending, err := queue.HasPendingUploads(ctx)
				if err != nil {
					t.Fatalf("Failed to check pending uploads: %v", err)
				}
				if !hasPending {
					t.Log("CRUD queue drained")
					goto drained
				}
			}
		}
	drained:

		// Stop upload loop
		uploadCancel()
		<-uploadDone

		// Verify database is still healthy after mutation
		if err := queue.CheckHealth(ctx); err != nil {
			t.Fatalf("Database unhealthy after mutation: %v", err)
		}
		t.Log("Database healthy after mutation")

		// Log bucket state for debugging
		var targetOp, lastOp int64
		err = db.QueryRow(ctx,
			"SELECT target_op, last_op FROM ps_buckets WHERE name = '$local'",
		).Scan(&targetOp, &lastOp)
		if err != nil {
			t.Logf("Could not read $local bucket state: %v", err)
		} else {
			t.Logf("$local bucket: target_op=%d, last_op=%d", targetOp, lastOp)
		}
	})
}
