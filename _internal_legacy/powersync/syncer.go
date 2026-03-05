package powersync

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	psapi "github.com/usetero/cli/internal/boundary/powersync"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync/db"
	"github.com/usetero/cli/internal/powersync/extension"
	"github.com/usetero/cli/internal/sqlite"
)

// TokenRefresher provides access tokens for authentication.
type TokenRefresher interface {
	GetAccessToken(ctx context.Context) (string, error)
	// ForceRefreshAccessToken refreshes the token unconditionally, bypassing
	// local expiration checks. Used when the server rejects a token the
	// client still considers valid (e.g. clock skew).
	ForceRefreshAccessToken(ctx context.Context) (string, error)
}

// Syncer manages PowerSync synchronization.
type Syncer interface {
	Start(ctx context.Context, db sqlite.DB, accountID string, onFirstSync func()) error
	Stop()
	State() State
	IsReady() bool
	NotifyUploadCompleted(ctx context.Context) error
}

// ControlPlane wraps extension control operations used by the syncer.
type ControlPlane interface {
	Start(ctx context.Context, req extension.StartRequest) ([]extension.Instruction, error)
	SendTextLine(ctx context.Context, line string) ([]extension.Instruction, error)
	NotifyConnection(ctx context.Context, event extension.ConnectionEvent) ([]extension.Instruction, error)
	NotifyTokenRefreshed(ctx context.Context) ([]extension.Instruction, error)
	NotifyUploadCompleted(ctx context.Context) ([]extension.Instruction, error)
	Close() error
}

var _ Syncer = (*syncer)(nil)
var _ ControlPlane = (*extension.Controller)(nil)

// syncer implements Syncer.
type syncer struct {
	endpoint       string
	tokenRefresher TokenRefresher
	scope          log.Scope
	clientFactory  func(endpoint string) psapi.Client
	controlPlaneFn func(db sqlite.DB) ControlPlane
	streamCapture  StreamCapture

	database    sqlite.DB
	accountID   string
	control     ControlPlane
	client      psapi.Client
	onFirstSync func()
	controlMu   sync.Mutex

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
func WithClientFactory(factory func(endpoint string) psapi.Client) SyncerOption {
	return func(s *syncer) {
		s.clientFactory = factory
	}
}

// WithControlPlaneFactory sets a custom control plane factory (for testing).
func WithControlPlaneFactory(factory func(db sqlite.DB) ControlPlane) SyncerOption {
	return func(s *syncer) {
		s.controlPlaneFn = factory
	}
}

// WithStreamCapture sets a best-effort raw sync-stream capture sink.
func WithStreamCapture(capture StreamCapture) SyncerOption {
	return func(s *syncer) {
		s.streamCapture = capture
	}
}

// NewSyncer creates a new Syncer. Call Start() when you have a database ready.
func NewSyncer(endpoint string, tokenRefresher TokenRefresher, scope log.Scope, opts ...SyncerOption) Syncer {
	s := &syncer{
		endpoint:       endpoint,
		tokenRefresher: tokenRefresher,
		scope:          scope.Child("powersync"),
		clientFactory:  psapi.NewClient,
		controlPlaneFn: func(db sqlite.DB) ControlPlane {
			return extension.NewController(db)
		},
	}
	for _, opt := range opts {
		opt(s)
	}
	s.setState(NewDisconnected())
	return s
}

// Start begins syncing. The onFirstSync callback fires once when initial sync completes.
func (s *syncer) Start(ctx context.Context, database sqlite.DB, accountID string, onFirstSync func()) error {
	if accountID == "" {
		return fmt.Errorf("powersync: accountID is required")
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
	s.control = s.controlPlaneFn(database)
	s.client = s.clientFactory(s.endpoint)
	s.client.SetToken(token)
	s.done = make(chan struct{})
	s.setState(NewConnecting())

	ctx, s.cancel = context.WithCancel(ctx)
	go s.run(ctx)

	s.scope.Info("sync started", log.String("accountID", accountID))
	return nil
}

// Stop shuts down syncing.
func (s *syncer) Stop() {
	if s.cancel != nil {
		s.cancel()
		<-s.done
		s.cancel = nil
	}
	s.controlMu.Lock()
	if s.control != nil {
		_ = s.control.Close()
		s.control = nil
	}
	s.controlMu.Unlock()

	if s.streamCapture != nil {
		if err := s.streamCapture.Close(); err != nil {
			s.scope.Warn("failed to close stream capture", "error", err)
		}
		s.streamCapture = nil
	}

	s.client = nil
	s.database = nil
	s.done = nil
	s.setState(NewDisconnected())
	s.scope.Info("sync stopped")
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

func (s *syncer) NotifyUploadCompleted(ctx context.Context) error {
	instructions, err := s.controlPlaneNotifyUploadCompleted(ctx)
	if errors.Is(err, errControlPlaneUnavailable) {
		// Upload completion can legitimately race with shutdown or run before start.
		return nil
	}
	if err != nil {
		return fmt.Errorf("notify upload completed: %w", err)
	}
	if _, err := s.applyInstructions(ctx, instructions); err != nil {
		return fmt.Errorf("apply upload completion instructions: %w", err)
	}
	return nil
}

func (s *syncer) setState(state State) {
	s.state.Store(&stateWrapper{state: state})
}
