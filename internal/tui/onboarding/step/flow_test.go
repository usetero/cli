package step

import (
	"errors"
	"testing"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/tui/keymap"
)

// testStep is a minimal Step implementation for testing
type testStep struct {
	closed    bool
	closeErr  error
	nextStep  Step
	nextErr   error
	nextCalls int
}

func (s *testStep) Init() tea.Cmd                  { return nil }
func (s *testStep) Update(tea.Msg) (Step, tea.Cmd) { return s, nil }
func (s *testStep) View() string                   { return "" }
func (s *testStep) SetSize(width, height int)      {}
func (s *testStep) IsBusy() bool                   { return false }
func (s *testStep) HasError() bool                 { return false }
func (s *testStep) Error() error                   { return nil }
func (s *testStep) Help() help.KeyMap              { return keymap.Simple{Keys: []key.Binding{}} }
func (s *testStep) Next() (Step, error) {
	s.nextCalls++
	return s.nextStep, s.nextErr
}
func (s *testStep) Close() error {
	s.closed = true
	return s.closeErr
}

func TestFlow_Close(t *testing.T) {
	t.Parallel()

	t.Run("closes current step", func(t *testing.T) {
		t.Parallel()
		logger := logtest.New(t)
		step := &testStep{nextErr: ErrNotReady}
		flow := NewFlow(step, logger)

		err := flow.Close()

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if !step.closed {
			t.Error("expected step to be closed")
		}
	})

	t.Run("returns error from step close", func(t *testing.T) {
		t.Parallel()
		logger := logtest.New(t)
		expectedErr := errors.New("close error")
		step := &testStep{nextErr: ErrNotReady, closeErr: expectedErr}
		flow := NewFlow(step, logger)

		err := flow.Close()

		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("returns nil when no current step", func(t *testing.T) {
		t.Parallel()
		logger := logtest.New(t)
		step := &testStep{} // Will complete immediately (nextStep=nil, nextErr=nil)
		flow := NewFlow(step, logger)

		// Trigger completion
		flow.Update(nil)

		err := flow.Close()

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func TestFlow_ClosesStepOnTransition(t *testing.T) {
	t.Parallel()

	t.Run("closes previous step when transitioning", func(t *testing.T) {
		t.Parallel()
		logger := logtest.New(t)

		step2 := &testStep{nextErr: ErrNotReady}
		step1 := &testStep{nextStep: step2}

		flow := NewFlow(step1, logger)

		// Trigger transition from step1 to step2
		flow.Update(nil)

		if !step1.closed {
			t.Error("expected step1 to be closed after transition")
		}
		if step2.closed {
			t.Error("expected step2 to not be closed yet")
		}
	})

	t.Run("closes final step when flow completes", func(t *testing.T) {
		t.Parallel()
		logger := logtest.New(t)

		step := &testStep{} // nextStep=nil, nextErr=nil means complete

		flow := NewFlow(step, logger)

		// Trigger completion
		flow.Update(nil)

		if !step.closed {
			t.Error("expected step to be closed after flow completes")
		}
		if !flow.IsComplete() {
			t.Error("expected flow to be complete")
		}
	})

	t.Run("logs error when step close fails on transition", func(t *testing.T) {
		t.Parallel()
		logger := logtest.New(t)

		step2 := &testStep{nextErr: ErrNotReady}
		step1 := &testStep{
			nextStep: step2,
			closeErr: errors.New("close failed"),
		}

		flow := NewFlow(step1, logger)

		// Trigger transition - should log error but continue
		flow.Update(nil)

		if !step1.closed {
			t.Error("expected step1 close to be called")
		}
		// Flow should continue despite close error
		if flow.IsComplete() {
			t.Error("expected flow to not be complete")
		}
	})
}
