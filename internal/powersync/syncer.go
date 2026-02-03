package powersync

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync/api"
	"github.com/usetero/cli/internal/powersync/db"
	"github.com/usetero/cli/internal/powersync/extension"
	"github.com/usetero/cli/internal/sqlite"
)

const (
	initialRetryDelay = 1 * time.Second
	maxRetryDelay     = 30 * time.Second
	maxAuthRetries    = 2
)

// TokenRefresher provides access tokens for authentication.
type TokenRefresher interface {
	GetAccessToken(ctx context.Context) (string, error)
}

// Syncer manages PowerSync synchronization.
type Syncer interface {
	Start(ctx context.Context, db sqlite.Database, accountID string, onFirstSync func()) error
	Stop()
	State() State
	IsReady() bool
}

var _ Syncer = (*syncer)(nil)

// syncer implements Syncer.
type syncer struct {
	endpoint       string
	tokenRefresher TokenRefresher
	log            log.Logger
	clientFactory  func(endpoint string) api.Client

	database    sqlite.Database
	accountID   string
	controller  *extension.Controller
	client      api.Client
	onFirstSync func()

	// State - protected by atomic operations
	state atomic.Pointer[stateWrapper]

	cancel context.CancelFunc
	done   chan struct{}
}

// stateWrapper wraps State to allow atomic.Pointer usage.
type stateWrapper struct {
	state State
}

// SyncerOption configures a Syncer.
type SyncerOption func(*syncer)

// WithClientFactory sets a custom client factory (for testing).
func WithClientFactory(factory func(endpoint string) api.Client) SyncerOption {
	return func(s *syncer) {
		s.clientFactory = factory
	}
}

// NewSyncer creates a new Syncer. Call Start() when you have a database ready.
func NewSyncer(endpoint string, tokenRefresher TokenRefresher, logger log.Logger, opts ...SyncerOption) Syncer {
	s := &syncer{
		endpoint:       endpoint,
		tokenRefresher: tokenRefresher,
		log:            logger,
		clientFactory:  api.NewClient,
	}
	for _, opt := range opts {
		opt(s)
	}
	s.setState(NewDisconnected())
	return s
}

// Start begins syncing. The onFirstSync callback fires once when initial sync completes.
func (s *syncer) Start(ctx context.Context, database sqlite.Database, accountID string, onFirstSync func()) error {
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

	if err := extension.ApplySchema(ctx, database); err != nil {
		return err
	}

	// Check database health before starting sync.
	// A crash during migration or write can leave the database in an
	// inconsistent state. If corrupt, return an error so caller can reset.
	queue := db.NewCrudQueue(database)
	if err := queue.CheckHealth(ctx); err != nil {
		return err
	}

	s.database = database
	s.accountID = accountID
	s.onFirstSync = onFirstSync
	s.controller = extension.NewController(database)
	s.client = s.clientFactory(s.endpoint)
	s.client.SetToken(token)
	s.done = make(chan struct{})
	s.setState(NewConnecting("Connecting..."))

	ctx, s.cancel = context.WithCancel(ctx)
	go s.run(ctx)

	s.log.Info("sync started", log.String("accountID", accountID))
	return nil
}

// Stop shuts down syncing.
func (s *syncer) Stop() {
	if s.cancel != nil {
		s.cancel()
		<-s.done
		s.cancel = nil
	}
	if s.controller != nil {
		s.controller.Close()
		s.controller = nil
	}
	s.client = nil
	s.database = nil
	s.done = nil
	s.setState(NewDisconnected())
	s.log.Info("sync stopped")
}

// State returns the current syncer state.
func (s *syncer) State() State {
	if w := s.state.Load(); w != nil {
		return w.state
	}
	return NewDisconnected()
}

// IsReady returns true if initial sync is complete.
func (s *syncer) IsReady() bool {
	_, ok := s.State().(*Ready)
	return ok
}

func (s *syncer) setState(state State) {
	s.state.Store(&stateWrapper{state: state})
}

// run is the main sync loop with retry logic.
func (s *syncer) run(ctx context.Context) {
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

		var clientErr *api.Error
		if errors.As(err, &clientErr) {
			if clientErr.IsAuth() {
				authRetries++
				if authRetries > maxAuthRetries {
					s.setError(fmt.Errorf("auth failed after %d retries: %w", maxAuthRetries, err))
					return
				}
				s.log.Debug("auth error, refreshing token", log.Int("attempt", authRetries))
				s.setState(NewSyncing("Reconnecting..."))
				if err := s.refreshToken(ctx); err != nil {
					s.setError(fmt.Errorf("token refresh: %w", err))
					return
				}
				continue
			}

			if clientErr.IsTransient() {
				s.log.Debug("transient error, retrying", log.Duration("delay", retryDelay), log.Any("error", err))
				s.setState(NewSyncing("Reconnecting..."))
				s.wait(ctx, retryDelay)
				retryDelay = min(retryDelay*2, maxRetryDelay)
				continue
			}
		}

		s.setError(err)
		return
	}
}

// syncOnce runs one sync session: start controller, connect stream, process lines.
func (s *syncer) syncOnce(ctx context.Context) error {
	instructions, err := s.controller.Start(ctx, extension.StartRequest{
		IncludeDefaults: true,
		Parameters:      map[string]any{"account_id": s.accountID},
	})
	if err != nil {
		return fmt.Errorf("controller start: %w", err)
	}

	for _, inst := range instructions {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		switch inst.Type {
		case extension.InstructionEstablishSyncStream:
			if err := s.connectAndSync(ctx, inst.Request); err != nil {
				return err
			}
		case extension.InstructionFetchCredentials:
			if err := s.refreshToken(ctx); err != nil {
				return err
			}
		default:
			// Other instruction types are handled elsewhere or ignored at this level
		}
	}

	return nil
}

// connectAndSync connects to the stream and processes lines until disconnect.
func (s *syncer) connectAndSync(ctx context.Context, req *api.SyncStreamRequest) error {
	if req == nil {
		return fmt.Errorf("no sync request")
	}

	s.log.Debug("connecting stream")
	s.setState(NewSyncing("Syncing your data..."))

	// Track if we've notified connection established
	connected := false

	err := s.client.SyncStream(ctx, req, func(line []byte) error {
		// Notify connection established on first line received
		// This ensures the PowerSync extension knows the stream is active
		if !connected {
			if _, err := s.controller.NotifyConnection(ctx, extension.ConnectionEstablished); err != nil {
				return fmt.Errorf("notify connected: %w", err)
			}
			connected = true
		}
		return s.handleLine(line)
	})

	if connected {
		_, _ = s.controller.NotifyConnection(ctx, extension.ConnectionEnded)
	}

	if err != nil {
		return fmt.Errorf("stream: %w", err)
	}
	return nil
}

// handleLine processes a single line from the sync stream.
func (s *syncer) handleLine(line []byte) error {
	ctx := context.Background() // Lines are processed synchronously

	instructions, err := s.controller.SendTextLine(ctx, string(line))
	if err != nil {
		return fmt.Errorf("send line: %w", err)
	}

	for _, inst := range instructions {
		switch inst.Type {
		case extension.InstructionDidCompleteSync:
			s.log.Debug("sync complete")
			s.setState(NewReady())
			s.fireFirstSync()

		case extension.InstructionUpdateSyncStatus:
			if inst.SyncStatus != nil && inst.SyncStatus.Downloading != nil {
				downloaded, total := inst.SyncStatus.Downloading.TotalProgress()
				if syncing, ok := s.State().(*Syncing); ok {
					s.setState(syncing.UpdateProgress(downloaded, total))
				} else {
					s.setState(NewSyncing("Syncing your data...").WithProgress(downloaded, total))
				}
				s.log.Debug("sync progress", "downloaded", downloaded, "total", total)
			}

		case extension.InstructionFetchCredentials:
			s.log.Debug("received FetchCredentials", "didExpire", inst.DidExpire)
			if err := s.refreshToken(ctx); err != nil {
				return err
			}

		case extension.InstructionCloseSyncStream:
			s.log.Debug("received CloseSyncStream")
			return nil

		case extension.InstructionLogLine:
			s.log.Debug("powersync", "severity", inst.Severity, "line", inst.Line)
			// Surface non-debug log lines as transient warnings in the UI
			// PowerSync sends important messages (like checkpoint issues) as INFO
			if inst.Severity != "debug" && inst.Line != "" {
				if syncing, ok := s.State().(*Syncing); ok {
					s.setState(syncing.WithWarning(inst.Line))
				}
			}
		default:
			// Other instruction types not expected during line processing
		}
	}

	return nil
}

func (s *syncer) fireFirstSync() {
	// Only fire once - check if we're transitioning to ready
	if s.onFirstSync != nil {
		s.log.Info("sync connected")
		s.onFirstSync()
		s.onFirstSync = nil // Don't fire again
	}
}

func (s *syncer) refreshToken(ctx context.Context) error {
	token, err := s.tokenRefresher.GetAccessToken(ctx)
	if err != nil {
		return err
	}
	s.client.SetToken(token)
	if _, err := s.controller.NotifyTokenRefreshed(ctx); err != nil {
		return fmt.Errorf("notify token refreshed: %w", err)
	}
	return nil
}

func (s *syncer) setError(err error) {
	s.setState(NewError(err))
	s.log.Error("sync failed", log.Any("error", err))
}

func (s *syncer) wait(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}
