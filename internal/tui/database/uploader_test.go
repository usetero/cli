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
