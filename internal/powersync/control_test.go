package powersync

import (
	"encoding/json"
	"testing"

	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
)

func TestInstruction_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("parses EstablishSyncStream", func(t *testing.T) {
		t.Parallel()

		data := `{"EstablishSyncStream": {"request": {"buckets": [], "client_id": "abc123"}}}`

		var inst Instruction
		if err := json.Unmarshal([]byte(data), &inst); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if inst.Type != InstructionEstablishSyncStream {
			t.Errorf("Type = %q, want %q", inst.Type, InstructionEstablishSyncStream)
		}
		if inst.Request == nil {
			t.Fatal("Request is nil")
		}
		if inst.Request.ClientID != "abc123" {
			t.Errorf("Request.ClientID = %q, want %q", inst.Request.ClientID, "abc123")
		}
	})

	t.Run("parses FetchCredentials", func(t *testing.T) {
		t.Parallel()

		data := `{"FetchCredentials": {"did_expire": true}}`

		var inst Instruction
		if err := json.Unmarshal([]byte(data), &inst); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if inst.Type != InstructionFetchCredentials {
			t.Errorf("Type = %q, want %q", inst.Type, InstructionFetchCredentials)
		}
		if inst.DidExpire == nil || !*inst.DidExpire {
			t.Error("DidExpire should be true")
		}
	})

	t.Run("parses CloseSyncStream", func(t *testing.T) {
		t.Parallel()

		data := `{"CloseSyncStream": {"hide_disconnect": false}}`

		var inst Instruction
		if err := json.Unmarshal([]byte(data), &inst); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if inst.Type != InstructionCloseSyncStream {
			t.Errorf("Type = %q, want %q", inst.Type, InstructionCloseSyncStream)
		}
		if inst.HideDisconnect == nil || *inst.HideDisconnect {
			t.Error("HideDisconnect should be false")
		}
	})

	t.Run("parses UpdateSyncStatus", func(t *testing.T) {
		t.Parallel()

		data := `{"UpdateSyncStatus": {"status": {"connected": true}}}`

		var inst Instruction
		if err := json.Unmarshal([]byte(data), &inst); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if inst.Type != InstructionUpdateSyncStatus {
			t.Errorf("Type = %q, want %q", inst.Type, InstructionUpdateSyncStatus)
		}
		if inst.Status == nil {
			t.Fatal("Status is nil")
		}
	})

	t.Run("parses LogLine", func(t *testing.T) {
		t.Parallel()

		data := `{"LogLine": {"severity": "info", "line": "test message"}}`

		var inst Instruction
		if err := json.Unmarshal([]byte(data), &inst); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if inst.Type != InstructionLogLine {
			t.Errorf("Type = %q, want %q", inst.Type, InstructionLogLine)
		}
		if inst.Severity != "info" {
			t.Errorf("Severity = %q, want %q", inst.Severity, "info")
		}
		if inst.Line != "test message" {
			t.Errorf("Line = %q, want %q", inst.Line, "test message")
		}
	})

	t.Run("parses DidCompleteSync", func(t *testing.T) {
		t.Parallel()

		data := `{"DidCompleteSync": {}}`

		var inst Instruction
		if err := json.Unmarshal([]byte(data), &inst); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if inst.Type != InstructionDidCompleteSync {
			t.Errorf("Type = %q, want %q", inst.Type, InstructionDidCompleteSync)
		}
	})

	t.Run("parses unknown instruction type", func(t *testing.T) {
		t.Parallel()

		data := `{"SomeNewInstruction": {"foo": "bar"}}`

		var inst Instruction
		if err := json.Unmarshal([]byte(data), &inst); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if inst.Type != "SomeNewInstruction" {
			t.Errorf("Type = %q, want %q", inst.Type, "SomeNewInstruction")
		}
	})
}

// setupTestDB creates a test database with the PowerSync extension loaded and schema initialized.
func setupTestDB(t *testing.T) sqlite.Database {
	t.Helper()

	db := sqlitetest.OpenTest(t)

	extPath, err := ExtensionPath()
	if err != nil {
		t.Fatalf("ExtensionPath() error = %v", err)
	}

	if err := db.LoadExtension(extPath, "sqlite3_powersync_init"); err != nil {
		t.Fatalf("LoadExtension() error = %v", err)
	}

	// Initialize schema - required before using powersync_control
	// Use a minimal test schema
	schemaJSON := `{"tables": [{"name": "test", "columns": [{"name": "id", "type": "text"}, {"name": "name", "type": "text"}]}]}`

	if _, err := db.Exec("SELECT powersync_replace_schema(?)", schemaJSON); err != nil {
		t.Fatalf("powersync_replace_schema() error = %v", err)
	}

	return db
}

func TestController_Start(t *testing.T) {
	t.Parallel()

	t.Run("returns instructions for sync stream", func(t *testing.T) {
		t.Parallel()

		db := setupTestDB(t)
		controller := NewController(db)

		// Act
		instructions, err := controller.Start(StartRequest{
			IncludeDefaults: true,
		})

		// Assert
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		// Should return instructions including EstablishSyncStream
		if len(instructions) == 0 {
			t.Fatal("expected at least one instruction")
		}

		// Look for EstablishSyncStream
		var found bool
		for _, inst := range instructions {
			if inst.Type == InstructionEstablishSyncStream {
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

		db := setupTestDB(t)
		controller := NewController(db)

		// Start first
		_, err := controller.Start(StartRequest{IncludeDefaults: true})
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		// Act
		_, err = controller.Stop()

		// Assert
		if err != nil {
			t.Errorf("Stop() error = %v", err)
		}
	})
}

func TestController_NotifyConnection(t *testing.T) {
	t.Parallel()

	t.Run("accepts established event", func(t *testing.T) {
		t.Parallel()

		db := setupTestDB(t)
		controller := NewController(db)

		// Start first
		_, _ = controller.Start(StartRequest{IncludeDefaults: true})

		// Act
		_, err := controller.NotifyConnection(ConnectionEstablished)

		// Assert
		if err != nil {
			t.Errorf("NotifyConnection(established) error = %v", err)
		}
	})

	t.Run("accepts end event", func(t *testing.T) {
		t.Parallel()

		db := setupTestDB(t)
		controller := NewController(db)

		// Start first
		_, _ = controller.Start(StartRequest{IncludeDefaults: true})

		// Act
		_, err := controller.NotifyConnection(ConnectionEnded)

		// Assert
		if err != nil {
			t.Errorf("NotifyConnection(end) error = %v", err)
		}
	})
}

func TestController_SendTextLine(t *testing.T) {
	t.Parallel()

	t.Run("processes JSON line from sync service", func(t *testing.T) {
		t.Parallel()

		db := setupTestDB(t)
		controller := NewController(db)

		// Start sync first
		_, _ = controller.Start(StartRequest{IncludeDefaults: true})
		_, _ = controller.NotifyConnection(ConnectionEstablished)

		// Act: send a checkpoint line (common sync message type)
		line := `{"checkpoint":{"last_op_id":"0","buckets":[]}}`
		instructions, err := controller.SendTextLine(line)

		// Assert: should process without error
		if err != nil {
			t.Errorf("SendTextLine() error = %v", err)
		}

		// May return instructions or nil depending on state
		_ = instructions
	})
}

func TestController_NotifyTokenRefreshed(t *testing.T) {
	t.Parallel()

	t.Run("notifies extension of refreshed token", func(t *testing.T) {
		t.Parallel()

		db := setupTestDB(t)
		controller := NewController(db)

		// Start first
		_, _ = controller.Start(StartRequest{IncludeDefaults: true})

		// Act
		_, err := controller.NotifyTokenRefreshed()

		// Assert
		if err != nil {
			t.Errorf("NotifyTokenRefreshed() error = %v", err)
		}
	})
}

// instructionTypes extracts the Type field from each instruction for error messages.
func instructionTypes(instructions []Instruction) []string {
	types := make([]string, len(instructions))
	for i, inst := range instructions {
		types[i] = inst.Type
	}
	return types
}
