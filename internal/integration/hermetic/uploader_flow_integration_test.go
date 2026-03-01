//go:build integration

package hermetic_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	graphapitest "github.com/usetero/cli/internal/boundary/graphql/apitest"
	psapitest "github.com/usetero/cli/internal/boundary/powersync/apitest"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	psdb "github.com/usetero/cli/internal/powersync/db"
	"github.com/usetero/cli/internal/powersync/db/dbtest"
	"github.com/usetero/cli/internal/powersync/powersynctest"
	"github.com/usetero/cli/internal/upload"
)

func TestIntegration_UploaderHermeticFlow(t *testing.T) {
	t.Parallel()

	database := dbtest.OpenTestDB(t)
	ctx := context.Background()

	_, err := database.Exec(ctx, "INSERT INTO ps_buckets (name, last_op, target_op) VALUES ('$local', 0, 0)")
	if err != nil {
		t.Fatalf("setup local bucket: %v", err)
	}

	convID := uuid.New().String()
	dbtest.InsertCrudEntry(t, database, 1, nil, `{"op":"PUT","type":"conversations","id":"`+convID+`","data":{"workspace_id":"ws-1","title":"Integration"}}`)

	var createdConversationID domain.ConversationID
	conversations := &graphapitest.MockConversations{
		CreateFunc: func(ctx context.Context, input graphql.CreateConversationInput) (*domain.Conversation, error) {
			createdConversationID = domain.ConversationID(input.ID.String())
			return &domain.Conversation{ID: domain.ConversationID(input.ID.String())}, nil
		},
	}

	uploader := upload.New(
		database,
		psapitest.NewMockClient(),
		powersynctest.NewMockTokenRefresher("token"),
		upload.MutationDeps{
			Conversations: conversations,
			Messages:      graphapitest.NewMockMessages(),
			Services:      graphapitest.NewMockAPIServiceServices(),
			Policies:      graphapitest.NewMockPolicies(),
		},
		logtest.NewScope(t),
	)

	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() {
		done <- uploader.Run(runCtx)
	}()

	select {
	case event := <-uploader.Events():
		if _, ok := event.(upload.SyncingEvent); !ok {
			t.Fatalf("expected SyncingEvent, got %T", event)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for sync event")
	}

	cancel()
	if err := <-done; err != nil && err != context.Canceled {
		t.Fatalf("uploader returned error: %v", err)
	}

	if createdConversationID == "" {
		t.Fatal("expected conversation mutation to be called")
	}

	queue := psdb.NewCrudQueue(database)
	entry, err := queue.GetNextEntry(ctx)
	if err != nil {
		t.Fatalf("GetNextEntry() error = %v", err)
	}
	if entry != nil {
		t.Fatal("expected CRUD queue to be empty after upload")
	}
}
