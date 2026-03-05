package syncer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/usetero/cli/internal/infrastructure/logging"
	psclient "github.com/usetero/cli/internal/infrastructure/powersync/client"
	psdb "github.com/usetero/cli/internal/infrastructure/powersync/db"
	"github.com/usetero/cli/internal/infrastructure/powersync/extension"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

var (
	errCloseSyncStream = errors.New("close sync stream")
)

// TokenSource supplies access tokens for PowerSync requests.
type TokenSource interface {
	GetAccessToken(ctx context.Context) (AccessToken, error)
	ForceRefreshAccessToken(ctx context.Context) (AccessToken, error)
}

// Client is the PowerSync stream HTTP client contract used by sync orchestration.
type Client interface {
	SyncStream(ctx context.Context, req *psclient.SyncStreamRequest, handler psclient.LineHandler) error
	SetToken(token psclient.AccessToken)
}

// ControlPlane is the PowerSync native control-plane contract.
type ControlPlane interface {
	Start(ctx context.Context, req extension.StartRequest) ([]extension.Instruction, error)
	SendTextLine(ctx context.Context, line string) ([]extension.Instruction, error)
	NotifyConnection(ctx context.Context, event extension.ConnectionEvent) ([]extension.Instruction, error)
	NotifyTokenRefreshed(ctx context.Context) ([]extension.Instruction, error)
	NotifyUploadCompleted(ctx context.Context) ([]extension.Instruction, error)
	Close() error
}

type clientFactory func(endpoint Endpoint) Client
type controlFactory func(db *sqlite.DB) ControlPlane

// Option configures Syncer.
type Option func(*Syncer)

// WithClientFactory overrides HTTP client creation (tests).
func WithClientFactory(factory func(endpoint Endpoint) Client) Option {
	return func(s *Syncer) {
		if factory != nil {
			s.clientFactory = factory
		}
	}
}

// WithControlFactory overrides control-plane creation (tests).
func WithControlFactory(factory func(db *sqlite.DB) ControlPlane) Option {
	return func(s *Syncer) {
		if factory != nil {
			s.controlFactory = factory
		}
	}
}

// WithRetryPolicy overrides reconnect policy.
func WithRetryPolicy(policy RetryPolicy) Option {
	return func(s *Syncer) {
		s.retry = policy
	}
}

// Syncer orchestrates PowerSync control-plane and stream lifecycle.
type Syncer struct {
	endpoint Endpoint
	tokens   TokenSource
	log      logging.Scope

	clientFactory  clientFactory
	controlFactory controlFactory

	retry RetryPolicy
	wait  func(ctx context.Context, d time.Duration)

	mu          sync.Mutex
	controlMu   sync.Mutex
	runCancel   context.CancelFunc
	runDone     chan struct{}
	db          *sqlite.DB
	client      Client
	control     ControlPlane
	accountID   AccountID
	onFirstSync func()
	firstSync   *sync.Once
	state       State
}

// New creates a new syncer.
func New(endpoint string, tokens TokenSource, log logging.Scope, opts ...Option) (*Syncer, error) {
	if err := extension.Register(); err != nil {
		return nil, fmt.Errorf("register powersync extension: %w", err)
	}

	parsedEndpoint, err := ParseEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	if tokens == nil {
		return nil, fmt.Errorf("%w: token source is required", ErrInvalidInput)
	}

	s := &Syncer{
		endpoint: parsedEndpoint,
		tokens:   tokens,
		log:      log.Child("powersync_syncer"),
		clientFactory: func(endpoint Endpoint) Client {
			return psclient.NewClient(endpoint.String())
		},
		controlFactory: func(db *sqlite.DB) ControlPlane {
			return extension.NewController(db)
		},
		retry: DefaultRetryPolicy(),
		wait: func(ctx context.Context, d time.Duration) {
			select {
			case <-ctx.Done():
			case <-time.After(d):
			}
		},
		state: &Disconnected{},
	}
	for _, opt := range opts {
		opt(s)
	}
	if err := s.retry.Validate(); err != nil {
		return nil, err
	}
	return s, nil
}

// Start applies schema, validates DB invariants, and starts sync loop.
func (s *Syncer) Start(ctx context.Context, db *sqlite.DB, accountID AccountID, onFirstSync func()) error {
	if db == nil {
		return fmt.Errorf("%w: database is required", ErrInvalidInput)
	}
	if err := accountID.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.runCancel != nil {
		return ErrAlreadyStarted
	}

	if err := extension.ApplySchema(ctx, db); err != nil {
		return fmt.Errorf("apply powersync schema: %w", err)
	}

	store := psdb.NewStore(db)
	if err := store.CheckHealth(ctx); err != nil {
		return fmt.Errorf("powersync db health check: %w", err)
	}

	token, err := s.tokens.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("get initial access token: %w", err)
	}

	s.db = db
	s.accountID = accountID
	s.onFirstSync = onFirstSync
	s.client = s.clientFactory(s.endpoint)
	s.client.SetToken(psclient.AccessToken(token))
	s.control = s.controlFactory(db)
	s.runDone = make(chan struct{})
	s.firstSync = &sync.Once{}
	s.setStateLocked(&Connecting{})

	runCtx, cancel := context.WithCancel(ctx)
	s.runCancel = cancel
	go s.run(runCtx, s.runDone)

	s.log.Info("sync started", "account_id", accountID)
	return nil
}

// Stop stops sync loop and closes control plane.
func (s *Syncer) Stop() {
	s.mu.Lock()
	cancel := s.runCancel
	done := s.runDone
	s.runCancel = nil
	s.runDone = nil
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}

	s.controlMu.Lock()
	defer s.controlMu.Unlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.control != nil {
		_ = s.control.Close()
		s.control = nil
	}
	s.client = nil
	s.db = nil
	s.accountID = ""
	s.onFirstSync = nil
	s.firstSync = nil
	s.setStateLocked(&Disconnected{})
}

// NotifyUploadCompleted notifies control plane that one upload batch completed.
func (s *Syncer) NotifyUploadCompleted(ctx context.Context) error {
	instructions, err := s.controlNotifyUploadCompleted(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrConnDone) {
			return nil
		}
		if errors.Is(err, errControlUnavailable) {
			return nil
		}
		if errors.Is(err, extension.ErrNoActiveIteration) {
			return nil
		}
		return fmt.Errorf("notify upload completed: %w", err)
	}
	_, err = s.applyInstructions(ctx, instructions)
	if err != nil {
		return fmt.Errorf("apply upload-complete instructions: %w", err)
	}
	return nil
}

// State returns current sync state.
func (s *Syncer) State() State {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state == nil {
		return &Disconnected{}
	}
	return s.state
}

// IsReady reports whether initial sync completed.
func (s *Syncer) IsReady() bool {
	_, ok := s.State().(*Ready)
	return ok
}

func (s *Syncer) setStateLocked(state State) {
	s.state = state
}

func (s *Syncer) setState(state State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.setStateLocked(state)
}

func (s *Syncer) takeRunDeps() (Client, ControlPlane, *sqlite.DB, AccountID, func(), *sync.Once) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.client, s.control, s.db, s.accountID, s.onFirstSync, s.firstSync
}

func (s *Syncer) clearFirstSyncCallback() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onFirstSync = nil
}
