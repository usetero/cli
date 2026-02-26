package powersync

import (
	"context"
	"errors"
	"fmt"

	"github.com/usetero/cli/internal/powersync/api"
	"github.com/usetero/cli/internal/powersync/extension"
)

var errCloseSyncStream = errors.New("close sync stream")

// runSession runs one sync session: start control plane, connect stream, process lines.
func (s *syncer) runSession(ctx context.Context) error {
	// Proactively refresh the token before connecting. This avoids a
	// needless 401 round-trip when the stream dropped after hours and the
	// token expired while we were connected. GetAccessToken checks the JWT
	// exp claim and only hits the network if the token is actually expired.
	// We only update the HTTP client here — no control plane notification,
	// since the control plane hasn't started its iteration yet.
	if token, err := s.tokenRefresher.GetAccessToken(ctx); err != nil {
		return err
	} else {
		s.client.SetToken(token)
	}

	instructions, err := s.controlPlaneStart(ctx, extension.StartRequest{
		IncludeDefaults: true,
		Parameters:      map[string]any{"account_id": s.accountID},
	})
	if err != nil {
		return fmt.Errorf("control plane start: %w", err)
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
		case extension.InstructionFetchCredentials:
			if err := s.refreshToken(ctx); err != nil {
				return err
			}
		default:
			// Other instruction types are handled elsewhere or ignored at this level.
		}
	}

	return nil
}

// runStream connects to the stream and processes lines until disconnect.
func (s *syncer) runStream(ctx context.Context, req *api.SyncStreamRequest) error {
	if req == nil {
		return fmt.Errorf("no sync request")
	}

	s.scope.Debug("connecting stream")
	s.setState(NewSyncing())

	connected := false

	err := s.client.SyncStream(ctx, req, func(line []byte) error {
		if !connected {
			if _, err := s.controlPlaneNotifyConnection(ctx, extension.ConnectionEstablished); err != nil {
				return fmt.Errorf("notify connected: %w", err)
			}
			connected = true
		}
		return s.processLine(ctx, line)
	})

	if connected {
		_, _ = s.controlPlaneNotifyConnection(ctx, extension.ConnectionEnded)
	}

	if err != nil {
		if errors.Is(err, errCloseSyncStream) {
			return nil
		}
		return fmt.Errorf("stream: %w", err)
	}
	return nil
}

// processLine handles one line from the sync stream.
func (s *syncer) processLine(ctx context.Context, line []byte) error {
	s.captureStreamLine(line)

	instructions, err := s.controlPlaneSendTextLine(ctx, string(line))
	if err != nil {
		return fmt.Errorf("send line: %w", err)
	}

	action, err := s.applyInstructions(ctx, instructions)
	if err != nil {
		return err
	}
	if action == streamActionClose {
		return errCloseSyncStream
	}
	return nil
}

func (s *syncer) captureStreamLine(line []byte) {
	if s.streamCapture == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			s.scope.Error("stream capture panicked; continuing sync", "panic", r)
		}
	}()
	s.streamCapture.CaptureLine(line)
}

func (s *syncer) fireFirstSync() {
	if s.onFirstSync != nil {
		s.scope.Info("sync connected")
		s.onFirstSync()
		s.onFirstSync = nil
	}
}
