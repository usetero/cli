package syncer

import (
	"context"
	"errors"
	"fmt"
	"time"

	psclient "github.com/usetero/cli/internal/infrastructure/powersync/client"
	"github.com/usetero/cli/internal/infrastructure/powersync/extension"
)

const controlParamAccountID = "account_id"

func (s *Syncer) run(ctx context.Context, done chan struct{}) {
	defer close(done)

	retries := 0
	delay := s.retry.InitialDelay

	for {
		if ctx.Err() != nil {
			return
		}

		err := s.runSession(ctx)
		if err == nil {
			retries = 0
			delay = s.retry.InitialDelay
			continue
		}
		if ctx.Err() != nil {
			return
		}

		var apiErr *psclient.Error
		if errors.As(err, &apiErr) {
			if apiErr.IsAuth() {
				s.transitionToReconnecting(retries)
				if refreshErr := s.forceRefreshToken(ctx); refreshErr != nil {
					retries++
					s.wait(ctx, delay)
					delay = minDuration(delay*2, s.retry.MaxDelay)
					continue
				}
				retries = 0
				delay = s.retry.InitialDelay
				continue
			}
			if apiErr.IsPermanent() {
				s.transitionToError(err)
				return
			}
		}

		retries++
		s.transitionToReconnecting(retries)
		s.wait(ctx, delay)
		delay = minDuration(delay*2, s.retry.MaxDelay)
	}
}

func (s *Syncer) runSession(ctx context.Context) error {
	deps := s.takeRunDeps()
	if deps.client == nil {
		return fmt.Errorf("syncer client unavailable")
	}

	// Refresh token lazily (only network if expired) before opening a new stream.
	token, err := s.tokens.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("get access token: %w", err)
	}
	deps.client.SetToken(psclient.AccessToken(token))

	instructions, err := s.controlStart(ctx, extension.StartRequest{
		IncludeDefaults: true,
		Parameters:      map[string]any{controlParamAccountID: deps.accountID.String()},
	})
	if err != nil {
		return fmt.Errorf("control start: %w", err)
	}

	for _, inst := range instructions {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		switch inst.Type {
		case extension.InstructionEstablishSyncStream:
			if err := s.runStream(ctx, inst.Request); err != nil {
				return err
			}
		default:
			_, err := s.applyInstructions(ctx, []extension.Instruction{inst})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Syncer) runStream(ctx context.Context, req *psclient.SyncStreamRequest) error {
	deps := s.takeRunDeps()
	if deps.client == nil {
		return fmt.Errorf("syncer client unavailable")
	}
	if req == nil {
		return fmt.Errorf("missing sync stream request")
	}

	s.transitionToSyncing(nil)
	connected := false

	err := deps.client.SyncStream(ctx, req, func(line []byte) error {
		if !connected {
			if _, err := s.controlNotifyConnection(ctx, extension.ConnectionEstablished); err != nil {
				return fmt.Errorf("notify connection established: %w", err)
			}
			connected = true
		}

		instructions, err := s.controlSendTextLine(ctx, string(line))
		if err != nil {
			return fmt.Errorf("control send text line: %w", err)
		}
		action, err := s.applyInstructions(ctx, instructions)
		if err != nil {
			return err
		}
		if action == streamActionClose {
			return errCloseSyncStream
		}
		return nil
	})

	if connected {
		if _, notifyErr := s.controlNotifyConnection(ctx, extension.ConnectionEnded); notifyErr != nil && !errors.Is(notifyErr, extension.ErrNoActiveIteration) {
			s.log.Debug("notify connection end failed", "error", notifyErr)
		}
	}

	if err != nil {
		if errors.Is(err, errCloseSyncStream) {
			return nil
		}
		return fmt.Errorf("sync stream: %w", err)
	}
	return nil
}

func (s *Syncer) applyInstructions(ctx context.Context, instructions []extension.Instruction) (streamAction, error) {
	for _, inst := range instructions {
		switch inst.Type {
		case extension.InstructionEstablishSyncStream:
			// Handled by runCycle before stream starts.
		case extension.InstructionDidCompleteSync:
			s.transitionToReady()
			s.fireFirstSync()
		case extension.InstructionUpdateSyncStatus:
			if inst.SyncStatus != nil && inst.SyncStatus.Downloading != nil {
				downloaded, total := inst.SyncStatus.Downloading.TotalProgress()
				s.transitionToSyncing(&Progress{Downloaded: downloaded, Total: total})
			}
		case extension.InstructionFetchCredentials:
			if err := s.refreshToken(ctx); err != nil {
				return streamActionContinue, err
			}
		case extension.InstructionCloseSyncStream:
			return streamActionClose, nil
		case extension.InstructionFlushFileSystem:
			deps := s.takeRunDeps()
			if deps.db == nil {
				return streamActionContinue, fmt.Errorf("database unavailable for flush")
			}
			if _, err := deps.db.Exec(ctx, "PRAGMA wal_checkpoint(PASSIVE)"); err != nil {
				return streamActionContinue, fmt.Errorf("flush filesystem: %w", err)
			}
		case extension.InstructionLogLine:
			s.log.Debug("powersync", "severity", inst.Severity, "line", inst.Line)
		}
	}
	return streamActionContinue, nil
}

func (s *Syncer) refreshToken(ctx context.Context) error {
	token, err := s.tokens.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("refresh access token: %w", err)
	}
	deps := s.takeRunDeps()
	if deps.client == nil {
		return fmt.Errorf("syncer client unavailable")
	}
	deps.client.SetToken(psclient.AccessToken(token))
	if _, err := s.controlNotifyTokenRefreshed(ctx); err != nil {
		return fmt.Errorf("notify token refreshed: %w", err)
	}
	return nil
}

func (s *Syncer) forceRefreshToken(ctx context.Context) error {
	token, err := s.tokens.ForceRefreshAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("force refresh access token: %w", err)
	}
	deps := s.takeRunDeps()
	if deps.client == nil {
		return fmt.Errorf("syncer client unavailable")
	}
	deps.client.SetToken(psclient.AccessToken(token))
	if _, err := s.controlNotifyTokenRefreshed(ctx); err != nil {
		return fmt.Errorf("notify token refreshed: %w", err)
	}
	return nil
}

func (s *Syncer) fireFirstSync() {
	deps := s.takeRunDeps()
	if deps.onFirstSync == nil || deps.firstSync == nil {
		return
	}
	deps.firstSync.Do(func() {
		deps.onFirstSync()
		s.clearFirstSyncCallback()
	})
}

type streamAction int

const (
	streamActionContinue streamAction = iota
	streamActionClose
)

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
