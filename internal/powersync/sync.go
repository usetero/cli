package powersync

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/usetero/cli/internal/sqlite"
)

// TokenRefresher provides access tokens for authentication.
type TokenRefresher interface {
	GetAccessToken(ctx context.Context) (string, error)
}

// Syncer manages data synchronization.
type Syncer interface {
	Stop()
	Status() Status
	LastError() error
	IsRunning() bool
}

// Ensure Sync implements Syncer.
var _ Syncer = (*Sync)(nil)

// Retry configuration.
const (
	initialRetryDelay = 1 * time.Second
	maxRetryDelay     = 30 * time.Second
	maxAuthRetries    = 2
)

// Sync manages PowerSync for a database.
//
// Call Start to begin syncing. The onFirstSync callback fires when the initial
// sync completes. Call Stop to shut down. Status and LastError are safe to call
// from any goroutine.
type Sync struct {
	// Immutable after construction
	config         *Config
	tokenRefresher TokenRefresher
	streamFactory  func(endpoint, token string) Streamer

	// Set by Start, read by background goroutine
	db          sqlite.Database
	accountID   string
	controller  *Controller
	stream      Streamer
	onFirstSync func()

	// Atomic state (safe for concurrent access)
	status    atomic.Value // Status
	lastError atomic.Value // error

	// Lifecycle
	firstSyncFired bool
	cancel         context.CancelFunc
	done           chan struct{}
}

// NewSync creates a Sync instance.
func NewSync(config *Config, tokenRefresher TokenRefresher) *Sync {
	s := &Sync{
		config:         config,
		tokenRefresher: tokenRefresher,
		streamFactory:  func(endpoint, token string) Streamer { return NewStream(endpoint, token) },
	}
	s.status.Store(StatusDisconnected)
	return s
}

// SetStreamFactory sets the stream factory. Must be called before Start.
// This is used by tests to inject mock streamers.
func (s *Sync) SetStreamFactory(factory func(endpoint, token string) Streamer) {
	s.streamFactory = factory
}

// Start begins syncing. The onFirstSync callback fires once when initial sync completes.
func (s *Sync) Start(ctx context.Context, db sqlite.Database, accountID string, onFirstSync func()) error {
	if accountID == "" {
		panic("powersync: accountID is required")
	}
	if s.cancel != nil {
		return fmt.Errorf("already started")
	}

	token, err := s.tokenRefresher.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("get initial token: %w", err)
	}

	if err := s.loadExtension(ctx, db); err != nil {
		return err
	}

	s.db = db
	s.accountID = accountID
	s.onFirstSync = onFirstSync
	s.firstSyncFired = false
	s.controller = NewController(db)
	s.stream = s.streamFactory(s.config.Endpoint, token)
	s.done = make(chan struct{})
	s.status.Store(StatusConnecting)

	ctx, s.cancel = context.WithCancel(ctx)
	go s.run(ctx)

	return nil
}

// Stop shuts down syncing. The database remains open.
func (s *Sync) Stop() {
	if s.cancel != nil {
		s.cancel()
		<-s.done
		s.cancel = nil
	}
	s.controller = nil
	s.stream = nil
	s.db = nil
	s.done = nil
	s.status.Store(StatusDisconnected)
}

// Status returns the current sync status.
func (s *Sync) Status() Status {
	if v := s.status.Load(); v != nil {
		return v.(Status)
	}
	return StatusDisconnected
}

// LastError returns the most recent error, or nil.
func (s *Sync) LastError() error {
	if v := s.lastError.Load(); v != nil {
		return v.(error)
	}
	return nil
}

// IsRunning returns true if sync is active.
func (s *Sync) IsRunning() bool {
	return s.cancel != nil
}

// --- Initialization ---

func (s *Sync) loadExtension(ctx context.Context, db sqlite.Database) error {
	extPath, err := ExtensionPath()
	if err != nil {
		return fmt.Errorf("get extension path: %w", err)
	}
	if err := db.LoadExtension(ctx, extPath, "sqlite3_powersync_init"); err != nil {
		return fmt.Errorf("load extension: %w", err)
	}
	if _, err := db.Exec(ctx, "SELECT powersync_replace_schema(?)", SchemaJSON()); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}

// --- Main Loop ---

func (s *Sync) run(ctx context.Context) {
	defer close(s.done)

	retryDelay := initialRetryDelay
	authRetries := 0

	for {
		if ctx.Err() != nil {
			return
		}

		err := s.syncOnce(ctx)
		if err == nil {
			retryDelay = initialRetryDelay
			authRetries = 0
			continue
		}
		if ctx.Err() != nil {
			return
		}

		// Handle error
		if IsAuthError(err) {
			authRetries++
			if authRetries > maxAuthRetries {
				s.fail(fmt.Errorf("auth failed after %d retries: %w", maxAuthRetries, err))
				return
			}
			s.status.Store(StatusReconnecting)
			if err := s.refreshToken(ctx); err != nil {
				s.fail(fmt.Errorf("token refresh: %w", err))
				return
			}
			continue
		}

		if IsTransientError(err) {
			s.status.Store(StatusReconnecting)
			s.lastError.Store(err)
			s.wait(ctx, retryDelay)
			retryDelay = min(retryDelay*2, maxRetryDelay)
			continue
		}

		// Permanent error
		s.fail(err)
		return
	}
}

func (s *Sync) syncOnce(ctx context.Context) error {
	instructions, err := s.controller.Start(ctx, StartRequest{
		IncludeDefaults: true,
		Parameters:      map[string]any{"account_id": s.accountID},
	})
	if err != nil {
		return fmt.Errorf("controller start: %w", err)
	}
	return s.handleInstructions(ctx, instructions)
}

func (s *Sync) wait(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}

func (s *Sync) fail(err error) {
	s.lastError.Store(err)
	s.status.Store(StatusError)
}

// --- Instruction Handling ---

func (s *Sync) handleInstructions(ctx context.Context, instructions []Instruction) error {
	for _, inst := range instructions {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := s.handleInstruction(ctx, inst); err != nil {
			return err
		}
	}
	return nil
}

func (s *Sync) handleInstruction(ctx context.Context, inst Instruction) error {
	switch inst.Type {
	case InstructionEstablishSyncStream:
		return s.connectStream(ctx, inst.Request)
	case InstructionFetchCredentials:
		return s.refreshToken(ctx)
	case InstructionCloseSyncStream:
		return nil // Stream closed, loop will reconnect
	case InstructionDidCompleteSync:
		s.status.Store(StatusConnected)
		s.fireFirstSync()
		return nil
	case InstructionUpdateSyncStatus:
		s.status.Store(StatusSyncing)
		return nil
	case InstructionLogLine:
		return nil // TODO: forward to logger
	default:
		return nil // Unknown instruction, ignore
	}
}

func (s *Sync) fireFirstSync() {
	if !s.firstSyncFired && s.onFirstSync != nil {
		s.firstSyncFired = true
		s.onFirstSync()
	}
}

// --- Stream ---

func (s *Sync) connectStream(ctx context.Context, req *StreamingSyncRequest) error {
	if req == nil {
		return fmt.Errorf("no sync request")
	}

	s.status.Store(StatusSyncing)

	if _, err := s.controller.NotifyConnection(ctx, ConnectionEstablished); err != nil {
		return fmt.Errorf("notify connected: %w", err)
	}

	err := s.stream.Connect(ctx, req, func(line []byte) error {
		instructions, err := s.controller.SendTextLine(ctx, string(line))
		if err != nil {
			return fmt.Errorf("send line: %w", err)
		}
		return s.handleInstructions(ctx, instructions)
	})

	// Always notify disconnection
	_, _ = s.controller.NotifyConnection(ctx, ConnectionEnded)

	if err != nil {
		return fmt.Errorf("stream: %w", err)
	}
	return nil
}

// --- Token ---

func (s *Sync) refreshToken(ctx context.Context) error {
	token, err := s.tokenRefresher.GetAccessToken(ctx)
	if err != nil {
		return err
	}
	s.stream.SetToken(token)
	if _, err := s.controller.NotifyTokenRefreshed(ctx); err != nil {
		return fmt.Errorf("notify token refreshed: %w", err)
	}
	return nil
}
