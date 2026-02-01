package powersync_test

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/powersync/powersynctest"
)

func TestController_Start(t *testing.T) {
	t.Parallel()

	t.Run("returns EstablishSyncStream instruction", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := powersynctest.OpenTestDB(t)
		controller := powersync.NewController(db)

		instructions, err := controller.Start(ctx, powersync.StartRequest{
			IncludeDefaults: true,
		})
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		if len(instructions) == 0 {
			t.Fatal("expected at least one instruction")
		}

		var found bool
		for _, inst := range instructions {
			if inst.Type == powersync.InstructionEstablishSyncStream {
				found = true
				if inst.Request == nil {
					t.Error("EstablishSyncStream should have a Request")
				}
			}
		}
		if !found {
			t.Errorf("expected EstablishSyncStream instruction, got: %v", instructionTypes(instructions))
		}
	})
}

func TestController_Stop(t *testing.T) {
	t.Parallel()

	t.Run("stops without error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := powersynctest.OpenTestDB(t)
		controller := powersync.NewController(db)

		_, err := controller.Start(ctx, powersync.StartRequest{IncludeDefaults: true})
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		_, err = controller.Stop(ctx)
		if err != nil {
			t.Errorf("Stop() error = %v", err)
		}
	})
}

func TestController_NotifyConnection(t *testing.T) {
	t.Parallel()

	t.Run("accepts established event", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := powersynctest.OpenTestDB(t)
		controller := powersync.NewController(db)

		_, _ = controller.Start(ctx, powersync.StartRequest{IncludeDefaults: true})

		_, err := controller.NotifyConnection(ctx, powersync.ConnectionEstablished)
		if err != nil {
			t.Errorf("NotifyConnection(established) error = %v", err)
		}
	})

	t.Run("accepts end event", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := powersynctest.OpenTestDB(t)
		controller := powersync.NewController(db)

		_, _ = controller.Start(ctx, powersync.StartRequest{IncludeDefaults: true})

		_, err := controller.NotifyConnection(ctx, powersync.ConnectionEnded)
		if err != nil {
			t.Errorf("NotifyConnection(end) error = %v", err)
		}
	})
}

func TestController_SendTextLine(t *testing.T) {
	t.Parallel()

	t.Run("processes checkpoint line", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := powersynctest.OpenTestDB(t)
		controller := powersync.NewController(db)

		_, _ = controller.Start(ctx, powersync.StartRequest{IncludeDefaults: true})
		_, _ = controller.NotifyConnection(ctx, powersync.ConnectionEstablished)

		line := `{"checkpoint":{"last_op_id":"0","buckets":[]}}`
		_, err := controller.SendTextLine(ctx, line)
		if err != nil {
			t.Errorf("SendTextLine() error = %v", err)
		}
	})
}

func TestController_NotifyTokenRefreshed(t *testing.T) {
	t.Parallel()

	t.Run("notifies without error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := powersynctest.OpenTestDB(t)
		controller := powersync.NewController(db)

		_, _ = controller.Start(ctx, powersync.StartRequest{IncludeDefaults: true})

		_, err := controller.NotifyTokenRefreshed(ctx)
		if err != nil {
			t.Errorf("NotifyTokenRefreshed() error = %v", err)
		}
	})
}

func instructionTypes(instructions []powersync.Instruction) []string {
	types := make([]string, len(instructions))
	for i, inst := range instructions {
		types[i] = inst.Type
	}
	return types
}
