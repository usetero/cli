package session

import (
	"context"
)

// Stop tears down account runtime lifecycle resources.
func (s *Service) Stop() error {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	if !s.state.Running {
		s.mu.Unlock()
		return nil
	}

	accountID := s.state.AccountID
	cancel := s.cancel
	syncer := s.syncer
	db := s.db

	s.state = State{}
	s.cancel = nil
	s.syncer = nil
	s.uploader = nil
	s.db = nil
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if syncer != nil {
		syncer.Stop()
	}
	s.wg.Wait()
	if db != nil {
		if err := db.Close(); err != nil {
			return err
		}
	}

	s.emit(context.Background(), Event{Kind: EventStopped, AccountID: accountID})
	s.log.Info("session stopped", "account_id", accountID)
	return nil
}
