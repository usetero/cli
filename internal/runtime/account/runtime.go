package account

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/usetero/cli/internal/infrastructure/logging"
	psclient "github.com/usetero/cli/internal/infrastructure/powersync/client"
	psdb "github.com/usetero/cli/internal/infrastructure/powersync/db"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	psuploader "github.com/usetero/cli/internal/infrastructure/powersync/uploader"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

// Config is the production account runtime configuration.
type Config struct {
	Env             string
	PowerSyncOrigin string
	SyncerTokens    pssyncer.TokenSource
	UploaderTokens  psuploader.TokenSource
}

// Runtime is one running account-local database runtime: sqlite + syncer + uploader.
type Runtime struct {
	scope  Scope
	dbPath sqlite.DatabasePath

	db       *sqlite.DB
	syncer   testableSyncer
	uploader testableUploader

	hasCompletedInitialSync bool

	cancel context.CancelFunc
	events chan Event
	wg     sync.WaitGroup

	log logging.Scope

	mu     sync.Mutex
	closed bool
}

func New(ctx context.Context, scope Scope, cfg Config, log logging.Scope) (*Runtime, error) {
	if cfg.Env == "" {
		return nil, fmt.Errorf("env is required")
	}
	if cfg.PowerSyncOrigin == "" {
		return nil, fmt.Errorf("powersync origin is required")
	}
	if cfg.SyncerTokens == nil {
		return nil, fmt.Errorf("syncer tokens are required")
	}
	if cfg.UploaderTokens == nil {
		return nil, fmt.Errorf("uploader tokens are required")
	}

	return newTestable(ctx, scope, testableDeps{
		openDB: func(ctx context.Context, path sqlite.DatabasePath) (*sqlite.DB, error) {
			return sqlite.Open(ctx, path.String())
		},
		pathForScope: func(scope Scope) (sqlite.DatabasePath, error) {
			storage, err := sqlite.NewDefaultStorage(cfg.Env, string(scope.Organization.ID))
			if err != nil {
				return "", err
			}
			return storage.DatabasePath(sqlite.AccountID(scope.Account.ID))
		},
		newSyncer: func() (testableSyncer, error) {
			return pssyncer.New(cfg.PowerSyncOrigin, cfg.SyncerTokens, log.Child("powersync_syncer"))
		},
		newUploader: func(db *sqlite.DB, notifier interface {
			NotifyUploadCompleted(ctx context.Context) error
		}) (testableUploader, error) {
			return psuploader.New(
				psdb.NewStore(db),
				psclient.NewClient(cfg.PowerSyncOrigin),
				cfg.UploaderTokens,
				log.Child("powersync_uploader"),
				psuploader.WithSyncNotifier(notifier),
			), nil
		},
	}, log)
}

func (r *Runtime) Close(ctx context.Context) error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}
	r.closed = true
	cancel := r.cancel
	syncer := r.syncer
	db := r.db
	scope := r.scope
	r.cancel = nil
	r.syncer = nil
	r.uploader = nil
	r.db = nil
	r.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if syncer != nil {
		syncer.Stop()
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		r.wg.Wait()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}

	if db != nil {
		if err := db.Close(); err != nil {
			return err
		}
	}

	r.emit(ctx, Event{Kind: EventStopped, Scope: scope})
	r.log.Info("account runtime stopped", "organization_id", scope.Organization.ID, "account_id", scope.Account.ID)
	return nil
}

func (r *Runtime) DB() *sqlite.DB {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.db
}

func isNonFatalStopErr(err error) bool {
	return err == nil || errors.Is(err, context.Canceled)
}
