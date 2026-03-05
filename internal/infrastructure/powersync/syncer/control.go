package syncer

import (
	"context"
	"errors"

	"github.com/usetero/cli/internal/infrastructure/powersync/extension"
)

var errControlUnavailable = errors.New("control plane unavailable")

func (s *Syncer) withControl(fn func(control ControlPlane) error) error {
	s.controlMu.Lock()
	defer s.controlMu.Unlock()

	s.mu.Lock()
	control := s.control
	s.mu.Unlock()
	if control == nil {
		return errControlUnavailable
	}
	return fn(control)
}

func (s *Syncer) controlStart(ctx context.Context, req extension.StartRequest) ([]extension.Instruction, error) {
	var instructions []extension.Instruction
	err := s.withControl(func(control ControlPlane) error {
		var err error
		instructions, err = control.Start(ctx, req)
		return err
	})
	return instructions, err
}

func (s *Syncer) controlSendTextLine(ctx context.Context, line string) ([]extension.Instruction, error) {
	var instructions []extension.Instruction
	err := s.withControl(func(control ControlPlane) error {
		var err error
		instructions, err = control.SendTextLine(ctx, line)
		return err
	})
	return instructions, err
}

func (s *Syncer) controlNotifyConnection(ctx context.Context, event extension.ConnectionEvent) ([]extension.Instruction, error) {
	var instructions []extension.Instruction
	err := s.withControl(func(control ControlPlane) error {
		var err error
		instructions, err = control.NotifyConnection(ctx, event)
		return err
	})
	return instructions, err
}

func (s *Syncer) controlNotifyTokenRefreshed(ctx context.Context) ([]extension.Instruction, error) {
	var instructions []extension.Instruction
	err := s.withControl(func(control ControlPlane) error {
		var err error
		instructions, err = control.NotifyTokenRefreshed(ctx)
		return err
	})
	return instructions, err
}

func (s *Syncer) controlNotifyUploadCompleted(ctx context.Context) ([]extension.Instruction, error) {
	var instructions []extension.Instruction
	err := s.withControl(func(control ControlPlane) error {
		var err error
		instructions, err = control.NotifyUploadCompleted(ctx)
		return err
	})
	return instructions, err
}
