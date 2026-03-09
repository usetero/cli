//go:build integration_live

package uploadlive_test

import (
	"context"
	"testing"
	"time"

	"github.com/usetero/cli/internal/auth"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	psapi "github.com/usetero/cli/internal/boundary/powersync"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/keyring"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
	psdb "github.com/usetero/cli/internal/powersync/db"
	"github.com/usetero/cli/internal/powersync/db/dbtest"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/upload"
	"github.com/usetero/cli/internal/workos"
)

// Live integration tests run against non-production services.
//
// Prerequisites:
//  1. task auth:login
//  2. task run (complete onboarding to set default account/workspace)
//  3. task test:integration:live
func TestIntegrationLive_Upload(t *testing.T) {
	ctx := context.Background()
	logger := logtest.NewScope(t)

	cliConfig := config.LoadCLIConfig()
	env := cliConfig.Environment()

	t.Logf("Environment: %s", env)
	t.Logf("API Origin: %s", cliConfig.APIOrigin)
	t.Logf("PowerSync Origin: %s", cliConfig.PowerSyncOrigin)

	storage := keyring.New(env)
	oauthProvider := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.ChatOrigin, cliConfig.PowerSyncOrigin)
	authSvc := auth.NewService(oauthProvider, storage, logger)

	if _, err := authSvc.GetAccessToken(ctx); err != nil {
		t.Fatalf("failed to get access token: %v (run: task auth:login)", err)
	}

	orgCfg, err := config.LoadOrgPreferences(env, config.ActiveOrgID(env))
	if err != nil {
		t.Fatalf("org preferences not found: %v (run: task run)", err)
	}
	orgPrefs := preferences.NewOrgService(orgCfg, logger)

	accountID := orgPrefs.GetDefaultAccountID()
	if accountID == "" {
		t.Fatalf("no default account (run: task run)")
	}

	workspaceID := orgPrefs.GetDefaultWorkspaceID()
	if workspaceID == "" {
		t.Fatalf("no default workspace (run: task run)")
	}

	t.Logf("Account ID: %s", accountID)
	t.Logf("Workspace ID: %s", workspaceID)

	services := graphql.NewServiceSet(cliConfig.APIOrigin+"/graphql", authSvc, logger).WithAccountID(accountID)

	t.Run("mutation round-trip maintains healthy database", func(t *testing.T) {
		database := dbtest.OpenTestDB(t)
		syncer := powersync.NewSyncer(cliConfig.PowerSyncOrigin, authSvc, logger)

		syncCtx, syncCancel := context.WithTimeout(ctx, 90*time.Second)
		defer syncCancel()

		firstSyncDone := make(chan struct{})
		err := syncer.Start(syncCtx, database, accountID.String(), func() {
			close(firstSyncDone)
		})
		if err != nil {
			t.Fatalf("failed to start sync: %v", err)
		}
		defer syncer.Stop()

		t.Log("waiting for initial sync")
		select {
		case <-firstSyncDone:
			t.Log("initial sync complete")
		case <-syncCtx.Done():
			t.Fatalf("timeout waiting for initial sync")
		}

		queue := psdb.NewCrudQueue(database)
		if err := queue.CheckHealth(ctx); err != nil {
			t.Fatalf("database unhealthy before mutation: %v", err)
		}

		uploader := upload.New(
			database,
			psapi.NewClient(cliConfig.PowerSyncOrigin),
			authSvc,
			upload.MutationDeps{
				Conversations: services.Conversations,
				Messages:      services.Messages,
				Services:      services.Services,
				Policies:      services.Policies,
			},
			logger,
			upload.WithBatchCompletedHook(func(hookCtx context.Context) error {
				return syncer.NotifyUploadCompleted(hookCtx)
			}),
		)

		uploadCtx, uploadCancel := context.WithCancel(ctx)
		defer uploadCancel()

		uploadDone := make(chan error, 1)
		go func() {
			uploadDone <- uploader.Run(uploadCtx)
		}()

		t.Log("creating conversation")
		convID, err := database.Conversations().Create(ctx, accountID, workspaceID)
		if err != nil {
			t.Fatalf("failed to create conversation: %v", err)
		}
		t.Logf("created conversation: %s", convID)

		t.Log("creating message")
		msgID, err := database.Messages().CreateUserMessage(ctx, accountID, convID, "Hello from integration live test")
		if err != nil {
			t.Fatalf("failed to create message: %v", err)
		}
		t.Logf("created message: %s", msgID)

		t.Log("waiting for CRUD queue to drain")
		drainTimeout := time.After(45 * time.Second)
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-drainTimeout:
				hasPending, _ := queue.HasPendingUploads(ctx)
				t.Fatalf("timeout waiting for CRUD queue to drain (hasPending=%v)", hasPending)
			case <-ticker.C:
				hasPending, err := queue.HasPendingUploads(ctx)
				if err != nil {
					t.Fatalf("failed to check pending uploads: %v", err)
				}
				if !hasPending {
					t.Log("CRUD queue drained")
					goto drained
				}
			}
		}
	drained:

		uploadCancel()
		if err := <-uploadDone; err != nil && err != context.Canceled {
			t.Fatalf("uploader stopped with error: %v", err)
		}

		if err := queue.CheckHealth(ctx); err != nil {
			t.Fatalf("database unhealthy after mutation: %v", err)
		}
	})
}
