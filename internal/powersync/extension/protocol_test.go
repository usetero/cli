package extension_test

import (
	"encoding/json"
	"testing"

	"github.com/usetero/cli/internal/powersync/extension"
)

func TestInstruction_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("parses EstablishSyncStream", func(t *testing.T) {
		t.Parallel()

		data := `{"EstablishSyncStream": {"request": {"buckets": [], "client_id": "abc123"}}}`

		var inst extension.Instruction
		if err := json.Unmarshal([]byte(data), &inst); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if inst.Type != extension.InstructionEstablishSyncStream {
			t.Errorf("Type = %q, want %q", inst.Type, extension.InstructionEstablishSyncStream)
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

		var inst extension.Instruction
		if err := json.Unmarshal([]byte(data), &inst); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if inst.Type != extension.InstructionFetchCredentials {
			t.Errorf("Type = %q, want %q", inst.Type, extension.InstructionFetchCredentials)
		}
		if inst.DidExpire == nil || !*inst.DidExpire {
			t.Error("DidExpire should be true")
		}
	})

	t.Run("parses CloseSyncStream", func(t *testing.T) {
		t.Parallel()

		data := `{"CloseSyncStream": {"hide_disconnect": false}}`

		var inst extension.Instruction
		if err := json.Unmarshal([]byte(data), &inst); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if inst.Type != extension.InstructionCloseSyncStream {
			t.Errorf("Type = %q, want %q", inst.Type, extension.InstructionCloseSyncStream)
		}
		if inst.HideDisconnect == nil || *inst.HideDisconnect {
			t.Error("HideDisconnect should be false")
		}
	})

	t.Run("parses UpdateSyncStatus with realistic payload", func(t *testing.T) {
		t.Parallel()

		// Realistic payload from the actual PowerSync extension
		data := `{"UpdateSyncStatus": {"status": {"connected": true, "connecting": false, "priority_status": [], "downloading": {"buckets": {"prio_3": {"priority": 3, "at_last": 0, "since_last": 1000, "target_count": 32920}}}, "streams": [{"name": "account_data", "parameters": null, "priority": 3, "active": true, "is_default": true, "has_explicit_subscription": false, "expires_at": null, "last_synced_at": null, "progress": {"total": 32920, "downloaded": 1000}}]}}}`

		var inst extension.Instruction
		if err := json.Unmarshal([]byte(data), &inst); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if inst.Type != extension.InstructionUpdateSyncStatus {
			t.Errorf("Type = %q, want %q", inst.Type, extension.InstructionUpdateSyncStatus)
		}
		if inst.SyncStatus == nil {
			t.Fatal("SyncStatus is nil")
		}
		if !inst.SyncStatus.Connected {
			t.Error("SyncStatus.Connected should be true")
		}
		if inst.SyncStatus.Downloading == nil {
			t.Fatal("SyncStatus.Downloading is nil")
		}
		downloaded, total := inst.SyncStatus.Downloading.TotalProgress()
		if downloaded != 1000 || total != 32920 {
			t.Errorf("TotalProgress() = (%d, %d), want (1000, 32920)", downloaded, total)
		}
	})

	t.Run("parses UpdateSyncStatus after sync complete", func(t *testing.T) {
		t.Parallel()

		// Status after sync completes - no downloading field
		data := `{"UpdateSyncStatus": {"status": {"connected": true, "connecting": false, "priority_status": [{"priority": 2147483647, "last_synced_at": 1738443993, "has_synced": true}], "downloading": null, "streams": []}}}`

		var inst extension.Instruction
		if err := json.Unmarshal([]byte(data), &inst); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if inst.SyncStatus == nil {
			t.Fatal("SyncStatus is nil")
		}
		if inst.SyncStatus.Downloading != nil {
			t.Error("SyncStatus.Downloading should be nil after sync complete")
		}
		if len(inst.SyncStatus.PriorityStatus) != 1 {
			t.Errorf("PriorityStatus length = %d, want 1", len(inst.SyncStatus.PriorityStatus))
		}
	})

	t.Run("parses LogLine", func(t *testing.T) {
		t.Parallel()

		data := `{"LogLine": {"severity": "info", "line": "test message"}}`

		var inst extension.Instruction
		if err := json.Unmarshal([]byte(data), &inst); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if inst.Type != extension.InstructionLogLine {
			t.Errorf("Type = %q, want %q", inst.Type, extension.InstructionLogLine)
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

		var inst extension.Instruction
		if err := json.Unmarshal([]byte(data), &inst); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if inst.Type != extension.InstructionDidCompleteSync {
			t.Errorf("Type = %q, want %q", inst.Type, extension.InstructionDidCompleteSync)
		}
	})

	t.Run("parses unknown instruction type", func(t *testing.T) {
		t.Parallel()

		data := `{"SomeNewInstruction": {"foo": "bar"}}`

		var inst extension.Instruction
		if err := json.Unmarshal([]byte(data), &inst); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if inst.Type != "SomeNewInstruction" {
			t.Errorf("Type = %q, want %q", inst.Type, "SomeNewInstruction")
		}
	})

	t.Run("rejects unknown fields in known instruction types", func(t *testing.T) {
		t.Parallel()

		// This should fail because "unknown_field" is not expected
		data := `{"FetchCredentials": {"did_expire": true, "unknown_field": "value"}}`

		var inst extension.Instruction
		err := json.Unmarshal([]byte(data), &inst)
		if err == nil {
			t.Fatal("Expected error for unknown field, got nil")
		}
	})
}
