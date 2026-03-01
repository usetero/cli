package powersync

import (
	"context"
	"fmt"
	"testing"

	psapi "github.com/usetero/cli/internal/boundary/powersync"
	"github.com/usetero/cli/internal/boundary/powersync/apitest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync/extension"
)

func TestSyncer_RunStream_CloseInstructionEndsStream(t *testing.T) {
	t.Parallel()

	cp := &streamTestControlPlane{
		sendTextInstructions: []extension.Instruction{
			{Type: extension.InstructionCloseSyncStream},
		},
	}
	mockClient := apitest.NewMockClient()
	mockClient.SyncStreamFunc = func(ctx context.Context, req *psapi.SyncStreamRequest, handler psapi.LineHandler) error {
		return handler([]byte(`{"x":1}`))
	}

	s := &syncer{
		client:  mockClient,
		control: cp,
		scope:   logtest.NewScope(t),
	}

	err := s.runStream(context.Background(), &psapi.SyncStreamRequest{})
	if err != nil {
		t.Fatalf("runStream() error = %v", err)
	}
	if cp.notifyConnectionCalls != 2 {
		t.Fatalf("NotifyConnection calls = %d, want 2 (established + end)", cp.notifyConnectionCalls)
	}
}

func TestSyncer_ProcessLine_InstructionErrorPropagates(t *testing.T) {
	t.Parallel()

	cp := &streamTestControlPlane{
		sendTextInstructions: []extension.Instruction{
			{Type: extension.InstructionFetchCredentials},
		},
	}
	s := &syncer{
		tokenRefresher: &stubTokenRefresher{err: fmt.Errorf("token unavailable")},
		client:         apitest.NewMockClient(),
		control:        cp,
		scope:          logtest.NewScope(t),
	}

	err := s.processLine(context.Background(), []byte(`{"token_expires_in":3600}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSyncer_NotifyUploadCompleted_ForwardsToControlPlane(t *testing.T) {
	t.Parallel()

	cp := &streamTestControlPlane{}
	s := &syncer{
		control: cp,
		scope:   logtest.NewScope(t),
	}

	if err := s.NotifyUploadCompleted(context.Background()); err != nil {
		t.Fatalf("NotifyUploadCompleted() error = %v", err)
	}
	if cp.notifyUploadCompletedCalls != 1 {
		t.Fatalf("NotifyUploadCompleted calls = %d, want 1", cp.notifyUploadCompletedCalls)
	}
}

func TestSyncer_NotifyUploadCompleted_AppliesReturnedInstructions(t *testing.T) {
	t.Parallel()

	cp := &streamTestControlPlane{
		notifyUploadInstructions: []extension.Instruction{
			{Type: extension.InstructionFetchCredentials},
		},
	}
	mockClient := apitest.NewMockClient()
	s := &syncer{
		tokenRefresher: &stubTokenRefresher{token: "refreshed-token"},
		client:         mockClient,
		control:        cp,
		scope:          logtest.NewScope(t),
	}

	if err := s.NotifyUploadCompleted(context.Background()); err != nil {
		t.Fatalf("NotifyUploadCompleted() error = %v", err)
	}
	if mockClient.Token != "refreshed-token" {
		t.Fatalf("token = %q, want %q", mockClient.Token, "refreshed-token")
	}
	if cp.notifyTokenRefreshedCalls != 1 {
		t.Fatalf("NotifyTokenRefreshed calls = %d, want 1", cp.notifyTokenRefreshedCalls)
	}
}

func TestSyncer_ProcessLine_CapturePanicIsolated(t *testing.T) {
	t.Parallel()

	cp := &streamTestControlPlane{}
	s := &syncer{
		control: cp,
		scope:   logtest.NewScope(t),
		streamCapture: &panicCapture{
			panicValue: "boom",
		},
	}

	if err := s.processLine(context.Background(), []byte(`{"token_expires_in":3600}`)); err != nil {
		t.Fatalf("processLine() error = %v", err)
	}
}

type streamTestControlPlane struct {
	sendTextInstructions       []extension.Instruction
	notifyUploadInstructions   []extension.Instruction
	notifyConnectionCalls      int
	notifyUploadCompletedCalls int
	notifyTokenRefreshedCalls  int
}

type panicCapture struct {
	panicValue any
}

func (p *panicCapture) CaptureLine([]byte) {
	panic(p.panicValue)
}

func (p *panicCapture) Close() error {
	return nil
}

func (s *streamTestControlPlane) Start(context.Context, extension.StartRequest) ([]extension.Instruction, error) {
	return nil, nil
}

func (s *streamTestControlPlane) SendTextLine(context.Context, string) ([]extension.Instruction, error) {
	return s.sendTextInstructions, nil
}

func (s *streamTestControlPlane) NotifyConnection(context.Context, extension.ConnectionEvent) ([]extension.Instruction, error) {
	s.notifyConnectionCalls++
	return nil, nil
}

func (s *streamTestControlPlane) NotifyTokenRefreshed(context.Context) ([]extension.Instruction, error) {
	s.notifyTokenRefreshedCalls++
	return nil, nil
}

func (s *streamTestControlPlane) NotifyUploadCompleted(context.Context) ([]extension.Instruction, error) {
	s.notifyUploadCompletedCalls++
	return s.notifyUploadInstructions, nil
}

func (s *streamTestControlPlane) Close() error {
	return nil
}
