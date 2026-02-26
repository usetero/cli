package powersync

import (
	"context"
	"errors"

	"github.com/usetero/cli/internal/powersync/extension"
)

var errControlPlaneUnavailable = errors.New("control plane unavailable")

func (s *syncer) withControlPlaneLocked(fn func(c ControlPlane) error) error {
	s.controlMu.Lock()
	defer s.controlMu.Unlock()
	if s.control == nil {
		return errControlPlaneUnavailable
	}
	return fn(s.control)
}

func (s *syncer) controlPlaneStart(ctx context.Context, req extension.StartRequest) ([]extension.Instruction, error) {
	var instructions []extension.Instruction
	err := s.withControlPlaneLocked(func(c ControlPlane) error {
		var err error
		instructions, err = c.Start(ctx, req)
		return err
	})
	return instructions, err
}

func (s *syncer) controlPlaneSendTextLine(ctx context.Context, line string) ([]extension.Instruction, error) {
	var instructions []extension.Instruction
	err := s.withControlPlaneLocked(func(c ControlPlane) error {
		var err error
		instructions, err = c.SendTextLine(ctx, line)
		return err
	})
	return instructions, err
}

func (s *syncer) controlPlaneNotifyConnection(ctx context.Context, event extension.ConnectionEvent) ([]extension.Instruction, error) {
	var instructions []extension.Instruction
	err := s.withControlPlaneLocked(func(c ControlPlane) error {
		var err error
		instructions, err = c.NotifyConnection(ctx, event)
		return err
	})
	return instructions, err
}

func (s *syncer) controlPlaneNotifyTokenRefreshed(ctx context.Context) ([]extension.Instruction, error) {
	var instructions []extension.Instruction
	err := s.withControlPlaneLocked(func(c ControlPlane) error {
		var err error
		instructions, err = c.NotifyTokenRefreshed(ctx)
		return err
	})
	return instructions, err
}

func (s *syncer) controlPlaneNotifyUploadCompleted(ctx context.Context) ([]extension.Instruction, error) {
	var instructions []extension.Instruction
	err := s.withControlPlaneLocked(func(c ControlPlane) error {
		var err error
		instructions, err = c.NotifyUploadCompleted(ctx)
		return err
	})
	return instructions, err
}
