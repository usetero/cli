package extension_test

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/powersync/db/dbtest"
	"github.com/usetero/cli/internal/powersync/extension"
)

func TestController_Start(t *testing.T) {
	t.Parallel()

	t.Run("returns EstablishSyncStream instruction", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := dbtest.OpenTestDB(t)
		controller := extension.NewController(db)

		instructions, err := controller.Start(ctx, extension.StartRequest{
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
			if inst.Type == extension.InstructionEstablishSyncStream {
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
		db := dbtest.OpenTestDB(t)
		controller := extension.NewController(db)

		_, err := controller.Start(ctx, extension.StartRequest{IncludeDefaults: true})
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
		db := dbtest.OpenTestDB(t)
		controller := extension.NewController(db)

		_, _ = controller.Start(ctx, extension.StartRequest{IncludeDefaults: true})

		_, err := controller.NotifyConnection(ctx, extension.ConnectionEstablished)
		if err != nil {
			t.Errorf("NotifyConnection(established) error = %v", err)
		}
	})

	t.Run("accepts end event", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := dbtest.OpenTestDB(t)
		controller := extension.NewController(db)

		_, _ = controller.Start(ctx, extension.StartRequest{IncludeDefaults: true})

		_, err := controller.NotifyConnection(ctx, extension.ConnectionEnded)
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
		db := dbtest.OpenTestDB(t)
		controller := extension.NewController(db)

		_, _ = controller.Start(ctx, extension.StartRequest{IncludeDefaults: true})
		_, _ = controller.NotifyConnection(ctx, extension.ConnectionEstablished)

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
		db := dbtest.OpenTestDB(t)
		controller := extension.NewController(db)

		_, _ = controller.Start(ctx, extension.StartRequest{IncludeDefaults: true})

		_, err := controller.NotifyTokenRefreshed(ctx)
		if err != nil {
			t.Errorf("NotifyTokenRefreshed() error = %v", err)
		}
	})
}

func instructionTypes(instructions []extension.Instruction) []extension.InstructionType {
	types := make([]extension.InstructionType, len(instructions))
	for i, inst := range instructions {
		types[i] = inst.Type
	}
	return types
}

func TestController_ConnectionStateConsistency(t *testing.T) {
	t.Parallel()

	t.Run("maintains state across multiple operations", func(t *testing.T) {
		t.Parallel()

		// This test verifies the fix for the "No iteration is active" bug.
		// The PowerSync extension maintains per-connection state:
		// - Start() begins an iteration on a connection
		// - NotifyConnection/SendTextLine must use the SAME connection
		// If connection pooling gives us different connections, we get the error.

		ctx := context.Background()
		db := dbtest.OpenTestDB(t)
		controller := extension.NewController(db)
		defer controller.Close()

		// Start an iteration
		_, err := controller.Start(ctx, extension.StartRequest{IncludeDefaults: true})
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		// These must use the same connection as Start(), or we get
		// "No iteration is active" error
		_, err = controller.NotifyConnection(ctx, extension.ConnectionEstablished)
		if err != nil {
			t.Fatalf("NotifyConnection(established) error = %v", err)
		}

		// Send a line - requires iteration to be active
		line := `{"token_expires_in":3600}`
		_, err = controller.SendTextLine(ctx, line)
		if err != nil {
			t.Fatalf("SendTextLine() error = %v", err)
		}

		// End the connection
		_, err = controller.NotifyConnection(ctx, extension.ConnectionEnded)
		if err != nil {
			t.Fatalf("NotifyConnection(end) error = %v", err)
		}
	})
}
