package session

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/domains/tenancy"
)

// Start initializes account runtime (db, syncer, uploader) for one account.
func (s *Service) Start(ctx context.Context, accountID tenancy.AccountID) error {
	if err := s.validateStart(accountID); err != nil {
		return err
	}

	s.mu.Lock()
	if s.state.Running {
		s.mu.Unlock()
		return fmt.Errorf("session is already running for account %s", s.state.AccountID)
	}
	s.mu.Unlock()

	path, err := s.storage.DatabasePath(toSQLiteAccountID(accountID))
	if err != nil {
		return err
	}
	db, err := s.openDB(ctx, path)
	if err != nil {
		return err
	}
	syncer, err := s.newSyncer()
	if err != nil {
		_ = db.Close()
		return err
	}

	runCtx, cancel := context.WithCancel(ctx)
	if err := syncer.Start(runCtx, db, toSyncerAccountID(accountID), func() {
		s.emit(runCtx, Event{Kind: EventReady, AccountID: accountID})
	}); err != nil {
		cancel()
		_ = db.Close()
		return err
	}

	uploader, err := s.newUploader(db, syncer)
	if err != nil {
		cancel()
		syncer.Stop()
		_ = db.Close()
		return err
	}

	s.mu.Lock()
	s.db = db
	s.syncer = syncer
	s.uploader = uploader
	s.cancel = cancel
	s.state = State{
		Running:   true,
		AccountID: accountID,
		DBPath:    path,
	}
	s.mu.Unlock()

	s.emit(runCtx, Event{Kind: EventStarting, AccountID: accountID})
	s.log.Info("session started", "account_id", accountID, "db_path", path)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		err := uploader.Run(runCtx)
		if !isNonFatalStopErr(err) {
			s.emit(runCtx, Event{Kind: EventError, AccountID: accountID, Err: err})
			s.log.Error("uploader stopped with error", "account_id", accountID, "error", err)
		}
	}()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.forwardUploaderEvents(runCtx, accountID, uploader.Events())
	}()

	return nil
}
