package session

import (
	"context"
	"io"
	"sync"

	"github.com/usetero/cli/internal/infrastructure/logging"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	psuploader "github.com/usetero/cli/internal/infrastructure/powersync/uploader"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

// Storage resolves account-scoped database paths.
type Storage interface {
	DatabasePath(accountID sqlite.AccountID) (sqlite.DatabasePath, error)
}

// Syncer controls account-scoped PowerSync lifecycle.
type Syncer interface {
	Start(ctx context.Context, db *sqlite.DB, accountID pssyncer.AccountID, onFirstSync func()) error
	Stop()
	State() pssyncer.State
	IsReady() bool
	NotifyUploadCompleted(ctx context.Context) error
}

// Uploader runs mutation upload loop and emits uploader events.
type Uploader interface {
	Run(ctx context.Context) error
	Events() <-chan psuploader.Event
}

type syncerFactory func() (Syncer, error)
type uploaderFactory func(db *sqlite.DB, notifier interface {
	NotifyUploadCompleted(ctx context.Context) error
}) (Uploader, error)
type dbOpenFunc func(ctx context.Context, path sqlite.DatabasePath) (*sqlite.DB, error)

// Service owns account runtime lifecycle: db + syncer + uploader.
type Service struct {
	storage     Storage
	newSyncer   syncerFactory
	newUploader uploaderFactory
	openDB      dbOpenFunc
	log         logging.Scope

	lifecycleMu sync.Mutex
	mu          sync.RWMutex
	state       State
	scope       Scope
	db          *sqlite.DB
	syncer      Syncer
	uploader    Uploader
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	events      chan Event
}

// NewService constructs the account runtime session service.
func NewService(storage Storage, newSyncer syncerFactory, newUploader uploaderFactory, log logging.Scope) *Service {
	if storage == nil {
		panic("session service requires storage")
	}
	if newSyncer == nil {
		panic("session service requires syncer factory")
	}
	if newUploader == nil {
		panic("session service requires uploader factory")
	}
	if log.Path() == "" {
		log = logging.RootScope(logging.NewWithWriter(io.Discard, logging.LevelInfo))
	}
	return &Service{
		storage:     storage,
		newSyncer:   newSyncer,
		newUploader: newUploader,
		openDB: func(ctx context.Context, path sqlite.DatabasePath) (*sqlite.DB, error) {
			return sqlite.Open(ctx, path.String())
		},
		log:    log.Child("runtime_session"),
		events: make(chan Event, 64),
	}
}

// Events returns lifecycle events for observers (non-blocking stream).
func (s *Service) Events() <-chan Event {
	return s.events
}

// State returns current session state.
func (s *Service) State() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *Service) setOpenDB(fn dbOpenFunc) {
	if fn == nil {
		return
	}
	s.openDB = fn
}
