package database

import (
	"testing"

	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/upload"
)

func TestUploader_Update(t *testing.T) {
	t.Parallel()

	t.Run("uploadEventMsg returns nil when uploader is nil", func(t *testing.T) {
		t.Parallel()

		u := &Uploader{logger: logtest.New(t)}

		// When uploader is nil, listenEvents returns nil
		cmd := u.Update(uploadEventMsg{event: upload.SyncingEvent{ProcessedCount: 5}})

		if cmd != nil {
			t.Error("expected nil command when uploader is nil")
		}
	})

	t.Run("uploadDoneMsg clears uploader", func(t *testing.T) {
		t.Parallel()

		u := &Uploader{logger: logtest.New(t)}
		u.uploader = &upload.Uploader{}

		u.Update(uploadDoneMsg{})

		if u.uploader != nil {
			t.Error("uploader should be nil after uploadDoneMsg")
		}
	})

	t.Run("MessageProcessingEvent emits UploadEventMsg", func(t *testing.T) {
		t.Parallel()

		u := &Uploader{logger: logtest.New(t)}

		event := upload.MessageProcessingEvent{
			ConversationID: "conv-123",
			UserMessageID:  "msg-42",
		}

		cmd := u.Update(uploadEventMsg{event: event})
		if cmd == nil {
			t.Fatal("expected command, got nil")
		}

		// Execute the batch command to get the individual commands
		msg := cmd()
		batchMsg, ok := msg.([]func() any)
		if !ok {
			// Single command case (uploader is nil, so listenEvents returns nil)
			// The batch has one command that returns UploadEventMsg
			result := msg
			uploadMsg, ok := result.(UploadEventMsg)
			if !ok {
				t.Fatalf("expected UploadEventMsg, got %T", result)
			}
			if uploadMsg.Event != event {
				t.Error("event not preserved in UploadEventMsg")
			}
			return
		}

		// Find the UploadEventMsg in the batch
		var found bool
		for _, fn := range batchMsg {
			if result := fn(); result != nil {
				if uploadMsg, ok := result.(UploadEventMsg); ok {
					found = true
					if uploadMsg.Event != event {
						t.Error("event not preserved in UploadEventMsg")
					}
				}
			}
		}
		if !found {
			t.Error("UploadEventMsg not found in batch")
		}
	})
}

func TestUploader_Wait(t *testing.T) {
	t.Parallel()

	t.Run("safe to call when done is nil", func(t *testing.T) {
		t.Parallel()

		u := &Uploader{}

		// Should not block or panic
		u.Wait()
	})

	t.Run("blocks until done channel closes", func(t *testing.T) {
		t.Parallel()

		u := &Uploader{
			done: make(chan struct{}),
		}

		// Close immediately
		close(u.done)

		// Should not block
		u.Wait()
	})
}
