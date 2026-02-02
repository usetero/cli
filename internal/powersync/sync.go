package powersync

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/usetero/cli/internal/log"
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
	Status() Status
	SyncStatus() *SyncStatus
	LastError() error
	IsRunning() bool
}

var _ Syncer = (*Sync)(nil)

// Sync implements Syncer.
type Sync struct {
	endpoint       string
	tokenRefresher TokenRefresher
	log            log.Logger
	clientFactory  func(endpoint string) Client

	db          sqlite.Database
	accountID   string
	controller  *Controller
	client      Client
	onFirstSync func()

	status     atomic.Value // Status
	syncStatus atomic.Value // *SyncStatus
	lastError  atomic.Value // error

	firstSyncFired bool
	cancel         context.CancelFunc
	done           chan struct{}
}

// NewSync creates a new Sync. Call Start() when you have a database ready.
func NewSync(endpoint string, tokenRefresher TokenRefresher, logger log.Logger) *Sync {
	s := &Sync{
		endpoint:       endpoint,
		tokenRefresher: tokenRefresher,
		log:            logger,
		clientFactory:  NewClient,
	}
	s.status.Store(StatusDisconnected)
	return s
}

// SetClientFactory sets a custom client factory (for testing).
func (s *Sync) SetClientFactory(factory func(endpoint string) Client) {
	s.clientFactory = factory
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

	if err := ApplySchema(ctx, db); err != nil {
		return err
	}

	// Check database health before starting sync.
	// A crash during migration or write can leave the database in an
	// inconsistent state. If corrupt, return an error so caller can reset.
	queue := NewCrudQueue(db)
	if err := queue.CheckHealth(ctx); err != nil {
		return err
	}

	s.db = db
	s.accountID = accountID
	s.onFirstSync = onFirstSync
	s.firstSyncFired = false
	s.controller = NewController(db)
	s.client = s.clientFactory(s.endpoint)
	s.client.SetToken(token)
	s.done = make(chan struct{})
	s.status.Store(StatusConnecting)

	ctx, s.cancel = context.WithCancel(ctx)
	go s.run(ctx)

	s.log.Info("sync started", log.String("accountID", accountID))
	return nil
}

// Stop shuts down syncing.
func (s *Sync) Stop() {
	if s.cancel != nil {
		s.cancel()
		<-s.done
		s.cancel = nil
	}
	s.controller = nil
	s.client = nil
	s.db = nil
	s.done = nil
	s.status.Store(StatusDisconnected)
	s.log.Info("sync stopped")
}

func (s *Sync) Status() Status {
	if v := s.status.Load(); v != nil {
		if status, ok := v.(Status); ok {
			return status
		}
	}
	return StatusDisconnected
}

func (s *Sync) SyncStatus() *SyncStatus {
	if v := s.syncStatus.Load(); v != nil {
		if ss, ok := v.(*SyncStatus); ok {
			return ss
		}
	}
	return nil
}

func (s *Sync) LastError() error {
	if v := s.lastError.Load(); v != nil {
		if err, ok := v.(error); ok {
			return err
		}
	}
	return nil
}

func (s *Sync) IsRunning() bool {
	return s.cancel != nil
}

// run is the main sync loop with retry logic.
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

		var clientErr *ClientError
		if errors.As(err, &clientErr) {
			if clientErr.IsAuth() {
				authRetries++
				if authRetries > maxAuthRetries {
					s.setError(fmt.Errorf("auth failed after %d retries: %w", maxAuthRetries, err))
					return
				}
				s.log.Debug("auth error, refreshing token", log.Int("attempt", authRetries))
				s.status.Store(StatusReconnecting)
				if err := s.refreshToken(ctx); err != nil {
					s.setError(fmt.Errorf("token refresh: %w", err))
					return
				}
				continue
			}

			if clientErr.IsTransient() {
				s.log.Debug("transient error, retrying", log.Duration("delay", retryDelay), log.Any("error", err))
				s.status.Store(StatusReconnecting)
				s.lastError.Store(err)
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
func (s *Sync) syncOnce(ctx context.Context) error {
	instructions, err := s.controller.Start(ctx, StartRequest{
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
		case InstructionEstablishSyncStream:
			if err := s.connectAndSync(ctx, inst.Request); err != nil {
				return err
			}
		case InstructionFetchCredentials:
			if err := s.refreshToken(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

// connectAndSync connects to the stream and processes lines until disconnect.
func (s *Sync) connectAndSync(ctx context.Context, req *SyncStreamRequest) error {
	if req == nil {
		return fmt.Errorf("no sync request")
	}

	s.log.Debug("connecting stream")
	s.status.Store(StatusSyncing)

	if _, err := s.controller.NotifyConnection(ctx, ConnectionEstablished); err != nil {
		return fmt.Errorf("notify connected: %w", err)
	}

	err := s.client.SyncStream(ctx, req, s.handleLine)

	_, _ = s.controller.NotifyConnection(ctx, ConnectionEnded)

	if err != nil {
		return fmt.Errorf("stream: %w", err)
	}
	return nil
}

// handleLine processes a single line from the sync stream.
func (s *Sync) handleLine(line []byte) error {
	ctx := context.Background() // Lines are processed synchronously

	instructions, err := s.controller.SendTextLine(ctx, string(line))
	if err != nil {
		return fmt.Errorf("send line: %w", err)
	}

	for _, inst := range instructions {
		switch inst.Type {
		case InstructionDidCompleteSync:
			s.log.Debug("sync complete")
			s.status.Store(StatusConnected)
			s.fireFirstSync()

		case InstructionUpdateSyncStatus:
			s.status.Store(StatusSyncing)
			if inst.SyncStatus != nil {
				s.syncStatus.Store(inst.SyncStatus)
				if inst.SyncStatus.Downloading != nil {
					downloaded, total := inst.SyncStatus.Downloading.TotalProgress()
					s.log.Debug("sync progress", "downloaded", downloaded, "total", total)
				}
			}

		case InstructionFetchCredentials:
			s.log.Debug("received FetchCredentials", "didExpire", inst.DidExpire)
			if err := s.refreshToken(ctx); err != nil {
				return err
			}

		case InstructionCloseSyncStream:
			s.log.Debug("received CloseSyncStream")
			return nil

		case InstructionLogLine:
			s.log.Debug("powersync", "severity", inst.Severity, "line", inst.Line)
		}
	}

	return nil
}

func (s *Sync) fireFirstSync() {
	if !s.firstSyncFired && s.onFirstSync != nil {
		s.firstSyncFired = true
		s.log.Info("sync connected")
		s.onFirstSync()
	}
}

func (s *Sync) refreshToken(ctx context.Context) error {
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

func (s *Sync) setError(err error) {
	s.lastError.Store(err)
	s.status.Store(StatusError)
	s.log.Error("sync failed", log.Any("error", err))
}

func (s *Sync) wait(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}
