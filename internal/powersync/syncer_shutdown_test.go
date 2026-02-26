package powersync_test

import (
	"context"
	"testing"
	"time"

	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/powersync/api"
	"github.com/usetero/cli/internal/powersync/api/apitest"
	"github.com/usetero/cli/internal/powersync/db/dbtest"
	"github.com/usetero/cli/internal/powersync/extension"
	"github.com/usetero/cli/internal/powersync/powersynctest"
	"github.com/usetero/cli/internal/sqlite"
)

func TestSyncer_Stop_UnblocksBlockedControlPlaneCalls(t *testing.T) {
	t.Parallel()

	db := dbtest.OpenTestDB(t)
	cp := &cancelAwareControlPlane{}
	mockClient := apitest.NewMockClient()
	mockClient.SyncStreamFunc = func(ctx context.Context, req *api.SyncStreamRequest, handler api.LineHandler) error {
		if err := handler([]byte(`{"token_expires_in":3600}`)); err != nil {
			return err
		}
		<-ctx.Done()
		return ctx.Err()
	}

	s := powersync.NewSyncer(
		"https://example.com",
		powersynctest.NewMockTokenRefresher("token"),
		logtest.NewScope(t),
		powersync.WithClientFactory(apitest.NewMockClientFactory(mockClient)),
		powersync.WithControlPlaneFactory(func(sqlite.DB) powersync.ControlPlane { return cp }),
	)

	ctx := context.Background()
	if err := s.Start(ctx, db, "account-123", nil); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	done := make(chan struct{})
	go func() {
		s.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Expected: Stop cancels context and blocked control-plane calls return.
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() timed out while control-plane call was blocked")
	}
}

type cancelAwareControlPlane struct{}

func (c *cancelAwareControlPlane) Start(context.Context, extension.StartRequest) ([]extension.Instruction, error) {
	return []extension.Instruction{{Type: extension.InstructionEstablishSyncStream, Request: &api.SyncStreamRequest{}}}, nil
}

func (c *cancelAwareControlPlane) SendTextLine(ctx context.Context, line string) ([]extension.Instruction, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (c *cancelAwareControlPlane) NotifyConnection(context.Context, extension.ConnectionEvent) ([]extension.Instruction, error) {
	return nil, nil
}

func (c *cancelAwareControlPlane) NotifyTokenRefreshed(context.Context) ([]extension.Instruction, error) {
	return nil, nil
}

func (c *cancelAwareControlPlane) NotifyUploadCompleted(context.Context) ([]extension.Instruction, error) {
	return nil, nil
}

func (c *cancelAwareControlPlane) Close() error {
	return nil
}

var _ powersync.ControlPlane = (*cancelAwareControlPlane)(nil)
