package session

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/usetero/cli/internal/domains/tenancy"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

// Start initializes account runtime (db, syncer, uploader) for one account.
func (s *Service) Start(ctx context.Context, accountID tenancy.AccountID) error {
	if err := validateStartAccountID(string(accountID)); err != nil {
		return err
	}
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	return s.startLocked(ctx, accountID)
}

func (s *Service) startLocked(ctx context.Context, accountID tenancy.AccountID) error {
	s.mu.Lock()
	if s.state.Running {
		s.mu.Unlock()
		return fmt.Errorf("session is already running for account %s", s.state.AccountID)
	}
	s.mu.Unlock()

	syncer, err := s.newSyncer()
	if err != nil {
		return err
	}
	path, err := s.storage.DatabasePath(toSQLiteAccountID(accountID))
	if err != nil {
		return err
	}
	db, err := s.openDB(ctx, path)
	if err != nil {
		return err
	}

	runCtx, cancel := context.WithCancel(ctx)
	if err := syncer.Start(runCtx, db, toSyncerAccountID(accountID), func() {
		s.emit(runCtx, Event{Kind: EventReady, AccountID: accountID})
	}); err != nil {
		cancel()
		_ = db.Close()
		if errors.Is(err, pssyncer.ErrApplySchema) {
			s.log.Info("resetting derived sqlite database after schema apply failure", "account_id", accountID, "db_path", path, "error", err)
			if resetErr := resetSQLiteFiles(path); resetErr != nil {
				return fmt.Errorf("reset sqlite after schema apply failure: %w", resetErr)
			}

			syncer, err = s.newSyncer()
			if err != nil {
				return err
			}
			db, openErr := s.openDB(ctx, path)
			if openErr != nil {
				return openErr
			}
			runCtx, cancel = context.WithCancel(ctx)
			retryErr := syncer.Start(runCtx, db, toSyncerAccountID(accountID), func() {
				s.emit(runCtx, Event{Kind: EventReady, AccountID: accountID})
			})
			if retryErr != nil {
				cancel()
				_ = db.Close()
				return retryErr
			}
		} else {
			return err
		}
	}

	uploader, err := s.newUploader(db, syncer)
	if err != nil {
		cancel()
		syncer.Stop()
		_ = db.Close()
		return err
	}

	s.mu.Lock()
	scope := s.scope
	scope.Account.ID = accountID
	s.scope = scope
	s.db = db
	s.syncer = syncer
	s.uploader = uploader
	s.cancel = cancel
	s.state = State{
		Running:        true,
		OrganizationID: scope.Organization.ID,
		AccountID:      accountID,
		DBPath:         path,
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

func resetSQLiteFiles(path sqlite.DatabasePath) error {
	targets := []string{
		path.String(),
		path.String() + "-wal",
		path.String() + "-shm",
	}
	for _, target := range targets {
		if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}
