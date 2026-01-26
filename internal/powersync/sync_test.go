package powersync

import (
	"errors"
	"testing"
)

func TestNewSync(t *testing.T) {
	t.Parallel()

	t.Run("initial status is disconnected", func(t *testing.T) {
		t.Parallel()

		sync := NewSync(&Config{Endpoint: "https://example.com"}, nil)

		if sync.Status() != StatusDisconnected {
			t.Errorf("Status() = %v, want %v", sync.Status(), StatusDisconnected)
		}
	})
}

func TestSync_Status(t *testing.T) {
	t.Parallel()

	t.Run("returns current status", func(t *testing.T) {
		t.Parallel()

		sync := NewSync(&Config{}, nil)

		if sync.Status() != StatusDisconnected {
			t.Errorf("initial Status() = %v, want %v", sync.Status(), StatusDisconnected)
		}

		sync.setStatus(StatusConnecting)
		if sync.Status() != StatusConnecting {
			t.Errorf("Status() = %v, want %v", sync.Status(), StatusConnecting)
		}

		sync.setStatus(StatusReconnecting)
		if sync.Status() != StatusReconnecting {
			t.Errorf("Status() = %v, want %v", sync.Status(), StatusReconnecting)
		}
	})
}

func TestSync_LastError(t *testing.T) {
	t.Parallel()

	t.Run("returns nil when no error", func(t *testing.T) {
		t.Parallel()

		sync := NewSync(&Config{}, nil)

		if sync.LastError() != nil {
			t.Error("LastError() should be nil initially")
		}
	})

	t.Run("returns error after setError", func(t *testing.T) {
		t.Parallel()

		sync := NewSync(&Config{}, nil)
		testErr := errors.New("test error")

		sync.setError(testErr)

		if sync.LastError() == nil {
			t.Error("LastError() should not be nil after setError")
		}
		if sync.Status() != StatusError {
			t.Errorf("Status() = %v, want %v", sync.Status(), StatusError)
		}
	})
}

func TestSync_IsRunning(t *testing.T) {
	t.Parallel()

	t.Run("false when not started", func(t *testing.T) {
		t.Parallel()

		sync := NewSync(&Config{}, nil)

		if sync.IsRunning() {
			t.Error("IsRunning() should be false before Start")
		}
	})
}

func TestSync_handleInstruction(t *testing.T) {
	t.Parallel()

	t.Run("UpdateSyncStatus sets status to syncing", func(t *testing.T) {
		t.Parallel()

		sync := NewSync(&Config{}, nil)
		err := sync.handleInstruction(nil, Instruction{Type: InstructionUpdateSyncStatus})

		if err != nil {
			t.Errorf("handleInstruction() error = %v", err)
		}
		if sync.Status() != StatusSyncing {
			t.Errorf("Status() = %v, want %v", sync.Status(), StatusSyncing)
		}
	})

	t.Run("DidCompleteSync sets status to connected", func(t *testing.T) {
		t.Parallel()

		sync := NewSync(&Config{}, nil)
		err := sync.handleInstruction(nil, Instruction{Type: InstructionDidCompleteSync})

		if err != nil {
			t.Errorf("handleInstruction() error = %v", err)
		}
		if sync.Status() != StatusConnected {
			t.Errorf("Status() = %v, want %v", sync.Status(), StatusConnected)
		}
	})

	t.Run("FetchCredentials without refresher returns error", func(t *testing.T) {
		t.Parallel()

		sync := NewSync(&Config{}, nil)
		err := sync.handleInstruction(nil, Instruction{Type: InstructionFetchCredentials})

		if err == nil {
			t.Error("expected error for missing token refresher")
		}
	})
}

func TestIsAuthError(t *testing.T) {
	t.Parallel()

	t.Run("true for auth StreamError", func(t *testing.T) {
		t.Parallel()

		err := &StreamError{Kind: ErrorKindAuth, StatusCode: 401}
		if !IsAuthError(err) {
			t.Error("IsAuthError should return true for ErrorKindAuth")
		}
	})

	t.Run("false for transient StreamError", func(t *testing.T) {
		t.Parallel()

		err := &StreamError{Kind: ErrorKindTransient, StatusCode: 500}
		if IsAuthError(err) {
			t.Error("IsAuthError should return false for ErrorKindTransient")
		}
	})

	t.Run("true for wrapped auth error", func(t *testing.T) {
		t.Parallel()

		streamErr := &StreamError{Kind: ErrorKindAuth, StatusCode: 401}
		wrappedErr := errors.Join(errors.New("context"), streamErr)
		if !IsAuthError(wrappedErr) {
			t.Error("IsAuthError should return true for wrapped auth errors")
		}
	})
}

func TestIsTransientError(t *testing.T) {
	t.Parallel()

	t.Run("true for transient StreamError", func(t *testing.T) {
		t.Parallel()

		err := &StreamError{Kind: ErrorKindTransient, StatusCode: 500}
		if !IsTransientError(err) {
			t.Error("IsTransientError should return true for ErrorKindTransient")
		}
	})

	t.Run("false for permanent StreamError", func(t *testing.T) {
		t.Parallel()

		err := &StreamError{Kind: ErrorKindPermanent, StatusCode: 400}
		if IsTransientError(err) {
			t.Error("IsTransientError should return false for ErrorKindPermanent")
		}
	})
}
