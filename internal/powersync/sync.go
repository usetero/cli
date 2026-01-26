package powersync

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/usetero/cli/internal/sqlite"
)

// TokenRefresher provides fresh access tokens when requested.
type TokenRefresher interface {
	GetAccessToken(ctx context.Context) (string, error)
}

// Backoff constants for retry logic.
const (
	initialBackoff = 1 * time.Second
	maxBackoff     = 30 * time.Second
)

// Sync manages the PowerSync connection for a database.
// It loads the extension, handles the sync stream, and processes instructions.
//
// Concurrency: Start() and Stop() are called from the main goroutine.
// The sync loop runs in a background goroutine. Status() and LastError()
// can be called from any goroutine and use atomic access.
type Sync struct {
	config         *Config
	tokenRefresher TokenRefresher

	// status and lastError are accessed atomically from multiple goroutines
	status    atomic.Value // stores Status
	lastError atomic.Value // stores error

	// These fields are set once in Start() before the sync goroutine launches,
	// and cleared in Stop() after the sync goroutine exits. No concurrent access.
	db         sqlite.Database
	accountID  string
	controller *Controller
	stream     *Stream
	cancel     context.CancelFunc
	done       chan struct{}
}

// NewSync creates a new PowerSync sync manager.
func NewSync(config *Config, tokenRefresher TokenRefresher) *Sync {
	s := &Sync{
		config:         config,
		tokenRefresher: tokenRefresher,
	}
	s.status.Store(StatusDisconnected)
	return s
}

// Start loads the PowerSync extension, initializes the schema, and starts syncing.
// Panics if accountID or token are empty - PowerSync requires both.
func (s *Sync) Start(ctx context.Context, db sqlite.Database, accountID, token string) error {
	if accountID == "" {
		panic("powersync: accountID is required")
	}
	if token == "" {
		panic("powersync: token is required")
	}
	if s.cancel != nil {
		return fmt.Errorf("already started")
	}

	s.db = db
	s.accountID = accountID
	s.status.Store(StatusConnecting)

	// Load the PowerSync extension into SQLite
	extPath, err := ExtensionPath()
	if err != nil {
		return fmt.Errorf("get extension path: %w", err)
	}

	if err := db.LoadExtension(extPath, "sqlite3_powersync_init"); err != nil {
		return fmt.Errorf("load extension: %w", err)
	}

	// Apply embedded schema to create views
	if _, err := db.Exec("SELECT powersync_replace_schema(?)", SchemaJSON()); err != nil {
		return fmt.Errorf("replace schema: %w", err)
	}

	// Create controller and stream
	s.controller = NewController(db)
	s.stream = NewStream(s.config.Endpoint, token)
	s.done = make(chan struct{})

	syncCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	// Start sync loop in background
	go s.syncLoop(syncCtx)

	return nil
}

// Stop stops syncing. The database remains open.
func (s *Sync) Stop() {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}

	// Wait for sync loop to finish
	if s.done != nil {
		<-s.done
	}

	s.controller = nil
	s.stream = nil
	s.db = nil
	s.done = nil
	s.status.Store(StatusDisconnected)
}

// Status returns the current sync status.
func (s *Sync) Status() Status {
	return s.status.Load().(Status)
}

// LastError returns the most recent sync error, or nil if none.
func (s *Sync) LastError() error {
	if err := s.lastError.Load(); err != nil {
		return err.(error)
	}
	return nil
}

// IsRunning returns true if sync is active.
func (s *Sync) IsRunning() bool {
	return s.cancel != nil
}

// WaitForFirstSync blocks until the first sync completes (data in ps_buckets).
// Returns an error if the context is cancelled or sync fails.
func (s *Sync) WaitForFirstSync(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Check for sync error
			if err := s.LastError(); err != nil {
				return err
			}

			// Check if ps_buckets has data
			if s.db != nil {
				var count int64
				err := s.db.QueryRow("SELECT COUNT(*) FROM ps_buckets").Scan(&count)
				if err == nil && count > 0 {
					return nil
				}
			}
		}
	}
}

func (s *Sync) setStatus(status Status) {
	s.status.Store(status)
}

func (s *Sync) setError(err error) {
	s.lastError.Store(err)
	s.status.Store(StatusError)
}

// syncLoop runs the main sync loop with error-aware retry logic.
func (s *Sync) syncLoop(ctx context.Context) {
	defer close(s.done)

	backoff := initialBackoff
	authRetries := 0
	const maxAuthRetries = 2

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := s.runSyncIteration(ctx)
		if err == nil {
			// Success - reset state
			backoff = initialBackoff
			authRetries = 0
			continue
		}

		if ctx.Err() != nil {
			return
		}

		// Handle error based on type
		switch {
		case IsAuthError(err):
			// Auth error - try to refresh token and retry immediately
			authRetries++
			if authRetries > maxAuthRetries {
				// Too many auth failures - surface error to user
				s.setError(fmt.Errorf("authentication failed after %d retries: %w", maxAuthRetries, err))
				return
			}

			s.setStatus(StatusReconnecting)
			if refreshErr := s.refreshToken(ctx); refreshErr != nil {
				s.setError(fmt.Errorf("token refresh failed: %w", refreshErr))
				return
			}
			// Retry immediately after token refresh
			continue

		case IsTransientError(err):
			// Transient error - show reconnecting status and retry with backoff
			s.setStatus(StatusReconnecting)
			s.lastError.Store(err)

		default:
			// Permanent error - surface to user and stop
			s.setError(err)
			return
		}

		// Wait with exponential backoff (for transient errors)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}

		// Increase backoff for next iteration, capped at max
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

// runSyncIteration runs one sync iteration.
func (s *Sync) runSyncIteration(ctx context.Context) error {
	if s.controller == nil || s.stream == nil {
		return fmt.Errorf("not initialized")
	}

	// Start sync and get instructions from extension
	instructions, err := s.controller.Start(StartRequest{
		IncludeDefaults: true,
		Parameters: map[string]any{
			"account_id": s.accountID,
		},
	})
	if err != nil {
		return fmt.Errorf("start: %w", err)
	}

	// Process instructions
	return s.processInstructions(ctx, instructions)
}

// processInstructions processes instructions from the extension.
func (s *Sync) processInstructions(ctx context.Context, instructions []Instruction) error {
	for _, inst := range instructions {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := s.handleInstruction(ctx, inst); err != nil {
			return err
		}
	}
	return nil
}

// handleInstruction handles a single instruction.
func (s *Sync) handleInstruction(ctx context.Context, inst Instruction) error {
	switch inst.Type {
	case InstructionEstablishSyncStream:
		return s.establishSyncStream(ctx, inst.Request)

	case InstructionFetchCredentials:
		return s.refreshToken(ctx)

	case InstructionCloseSyncStream:
		// Stream will be closed, loop will reconnect
		return nil

	case InstructionDidCompleteSync:
		s.setStatus(StatusConnected)
		return nil

	case InstructionUpdateSyncStatus:
		s.setStatus(StatusSyncing)
		return nil

	case InstructionLogLine:
		// TODO: Forward to logger
		return nil

	default:
		// Unknown instruction, ignore
		return nil
	}
}

// refreshToken gets a fresh token and updates the stream.
func (s *Sync) refreshToken(ctx context.Context) error {
	if s.tokenRefresher == nil {
		return fmt.Errorf("no token refresher configured")
	}

	token, err := s.tokenRefresher.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("refresh token: %w", err)
	}

	s.stream.SetToken(token)

	// Notify extension that token was refreshed
	if _, err := s.controller.NotifyTokenRefreshed(); err != nil {
		return fmt.Errorf("notify token refreshed: %w", err)
	}

	return nil
}

// establishSyncStream connects to the sync service and streams data.
func (s *Sync) establishSyncStream(ctx context.Context, req *StreamingSyncRequest) error {
	if req == nil {
		return fmt.Errorf("no sync request provided")
	}

	s.setStatus(StatusSyncing)

	// Notify extension that connection is established
	if _, err := s.controller.NotifyConnection(ConnectionEstablished); err != nil {
		return fmt.Errorf("notify connection established: %w", err)
	}

	// Connect and process stream
	err := s.stream.Connect(ctx, req, func(line []byte) error {
		// Forward line to extension
		instructions, err := s.controller.SendTextLine(string(line))
		if err != nil {
			return fmt.Errorf("process line: %w", err)
		}

		// Process any resulting instructions
		return s.processInstructions(ctx, instructions)
	})

	// Notify extension that connection ended
	if _, notifyErr := s.controller.NotifyConnection(ConnectionEnded); notifyErr != nil {
		// Log but don't override the original error
		_ = notifyErr
	}

	if err != nil {
		return fmt.Errorf("stream: %w", err)
	}

	return nil
}
