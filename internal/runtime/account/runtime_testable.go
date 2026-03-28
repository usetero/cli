package account

import (
	"context"
	"fmt"
	"io"

	"github.com/usetero/cli/internal/infrastructure/logging"
	psdb "github.com/usetero/cli/internal/infrastructure/powersync/db"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	psuploader "github.com/usetero/cli/internal/infrastructure/powersync/uploader"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

type testableSyncer interface {
	Start(ctx context.Context, db *sqlite.DB, accountID pssyncer.AccountID, onFirstSync func()) error
	Stop()
	State() pssyncer.State
	IsReady() bool
	NotifyUploadCompleted(ctx context.Context) error
}

type testableUploader interface {
	Run(ctx context.Context) error
	Events() <-chan psuploader.Event
}

type testableDeps struct {
	pathForScope func(scope Scope) (sqlite.DatabasePath, error)
	openDB       func(ctx context.Context, path sqlite.DatabasePath) (*sqlite.DB, error)
	newSyncer    func() (testableSyncer, error)
	newUploader  func(db *sqlite.DB, notifier interface {
		NotifyUploadCompleted(ctx context.Context) error
	}) (testableUploader, error)
}

func newTestable(ctx context.Context, scope Scope, deps testableDeps, log logging.Scope) (*Runtime, error) {
	if err := scope.Validate(); err != nil {
		return nil, err
	}
	if deps.pathForScope == nil {
		return nil, fmt.Errorf("database path resolver is required")
	}
	if deps.openDB == nil {
		return nil, fmt.Errorf("database opener is required")
	}
	if deps.newSyncer == nil {
		return nil, fmt.Errorf("syncer constructor is required")
	}
	if deps.newUploader == nil {
		return nil, fmt.Errorf("uploader constructor is required")
	}
	if log.Path() == "" {
		log = logging.RootScope(logging.NewWithWriter(io.Discard, logging.LevelInfo))
	}

	path, err := deps.pathForScope(scope)
	if err != nil {
		return nil, err
	}

	db, err := deps.openDB(ctx, path)
	if err != nil {
		return nil, err
	}
	cleanupDB := true
	defer func() {
		if cleanupDB {
			_ = db.Close()
		}
	}()

	store := psdb.NewStore(db)
	hasCompletedInitialSync, err := store.HasCompletedInitialSync(ctx)
	if err != nil {
		return nil, err
	}

	syncer, err := deps.newSyncer()
	if err != nil {
		return nil, err
	}

	runCtx, cancel := context.WithCancel(ctx)
	cleanupCancel := true
	defer func() {
		if cleanupCancel {
			cancel()
		}
	}()

	r := &Runtime{
		scope:                   scope,
		dbPath:                  path,
		db:                      db,
		syncer:                  syncer,
		cancel:                  cancel,
		events:                  make(chan Event, 64),
		log:                     log.Child("account_runtime"),
		hasCompletedInitialSync: hasCompletedInitialSync,
	}

	r.emit(runCtx, Event{Kind: EventStarting, Scope: scope})

	if err := syncer.Start(runCtx, db, pssyncer.AccountID(scope.Account.ID), func() {
		if err := store.MarkInitialSyncComplete(runCtx); err != nil {
			r.log.Error("persist initial sync completion", "account_id", scope.Account.ID, "error", err)
		}
		r.mu.Lock()
		r.hasCompletedInitialSync = true
		r.mu.Unlock()
		r.emit(runCtx, Event{Kind: EventReady, Scope: scope})
	}); err != nil {
		return nil, err
	}
	cleanupSyncer := true
	defer func() {
		if cleanupSyncer {
			syncer.Stop()
		}
	}()

	uploader, err := deps.newUploader(db, syncer)
	if err != nil {
		return nil, err
	}
	r.uploader = uploader

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		err := uploader.Run(runCtx)
		if !isNonFatalStopErr(err) {
			r.emit(runCtx, Event{Kind: EventError, Scope: scope, Err: err})
			r.log.Error("uploader stopped with error", "account_id", scope.Account.ID, "error", err)
		}
	}()

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		r.forwardUploaderEvents(runCtx, uploader.Events())
	}()

	cleanupDB = false
	cleanupCancel = false
	cleanupSyncer = false

	r.log.Info("account runtime started", "organization_id", scope.Organization.ID, "account_id", scope.Account.ID, "db_path", path)
	return r, nil
}
