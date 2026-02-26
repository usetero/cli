package powersync

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/powersync/extension"
)

type streamAction int

const (
	streamActionContinue streamAction = iota
	streamActionClose
)

func (s *syncer) applyInstructions(ctx context.Context, instructions []extension.Instruction) (streamAction, error) {
	for _, inst := range instructions {
		switch inst.Type {
		case extension.InstructionDidCompleteSync:
			s.scope.Debug("sync complete")
			s.setState(NewReady())
			s.fireFirstSync()

		case extension.InstructionUpdateSyncStatus:
			if inst.SyncStatus != nil && inst.SyncStatus.Downloading != nil {
				downloaded, total := inst.SyncStatus.Downloading.TotalProgress()
				s.setState(NewSyncing().WithProgress(downloaded, total))
				s.scope.Debug("sync progress", "downloaded", downloaded, "total", total)
			}

		case extension.InstructionFetchCredentials:
			s.scope.Debug("received FetchCredentials", "didExpire", inst.DidExpire)
			if err := s.refreshToken(ctx); err != nil {
				return streamActionContinue, err
			}

		case extension.InstructionCloseSyncStream:
			s.scope.Debug("received CloseSyncStream")
			return streamActionClose, nil

		case extension.InstructionFlushFileSystem:
			// Native SQLite handles durability on commit. Treat this as a best-effort
			// checkpoint request to reduce WAL growth and honor the extension signal.
			if _, err := s.database.Exec(ctx, "PRAGMA wal_checkpoint(PASSIVE)"); err != nil {
				return streamActionContinue, fmt.Errorf("flush file system: %w", err)
			}

		case extension.InstructionLogLine:
			s.scope.Debug("powersync", "severity", inst.Severity, "line", inst.Line)
		default:
			// Other instruction types not expected during line processing.
		}
	}

	return streamActionContinue, nil
}
