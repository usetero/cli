package powersync

import (
	"context"
	"fmt"
	"testing"

	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync/api/apitest"
	"github.com/usetero/cli/internal/powersync/db/dbtest"
	"github.com/usetero/cli/internal/powersync/extension"
)

func TestSyncer_ApplyInstructions_CloseSyncStream(t *testing.T) {
	t.Parallel()

	s := &syncer{
		scope: logtest.NewScope(t),
	}

	action, err := s.applyInstructions(context.Background(), []extension.Instruction{
		{Type: extension.InstructionCloseSyncStream},
	})
	if err != nil {
		t.Fatalf("applyInstructions() error = %v", err)
	}
	if action != streamActionClose {
		t.Fatalf("action = %v, want %v", action, streamActionClose)
	}
}

func TestSyncer_ApplyInstructions_FlushFileSystem(t *testing.T) {
	t.Parallel()

	db := dbtest.OpenTestDB(t)
	s := &syncer{
		database: db,
		scope:    logtest.NewScope(t),
	}

	action, err := s.applyInstructions(context.Background(), []extension.Instruction{
		{Type: extension.InstructionFlushFileSystem},
	})
	if err != nil {
		t.Fatalf("applyInstructions() error = %v", err)
	}
	if action != streamActionContinue {
		t.Fatalf("action = %v, want %v", action, streamActionContinue)
	}
}

func TestSyncer_ApplyInstructions_FetchCredentialsThenClose(t *testing.T) {
	t.Parallel()

	mockClient := apitest.NewMockClient()
	ctrl := &stubController{}
	s := &syncer{
		tokenRefresher: &stubTokenRefresher{token: "new-token"},
		client:         mockClient,
		control:        ctrl,
		scope:          logtest.NewScope(t),
	}

	action, err := s.applyInstructions(context.Background(), []extension.Instruction{
		{Type: extension.InstructionFetchCredentials},
		{Type: extension.InstructionCloseSyncStream},
	})
	if err != nil {
		t.Fatalf("applyInstructions() error = %v", err)
	}
	if action != streamActionClose {
		t.Fatalf("action = %v, want %v", action, streamActionClose)
	}
	if mockClient.Token != "new-token" {
		t.Fatalf("token = %q, want %q", mockClient.Token, "new-token")
	}
	if ctrl.notifyTokenRefreshedCalls != 1 {
		t.Fatalf("NotifyTokenRefreshed calls = %d, want 1", ctrl.notifyTokenRefreshedCalls)
	}
}

func TestSyncer_ApplyInstructions_FetchCredentialsError(t *testing.T) {
	t.Parallel()

	mockClient := apitest.NewMockClient()
	ctrl := &stubController{
		notifyTokenRefreshedErr: fmt.Errorf("notify failed"),
	}
	s := &syncer{
		tokenRefresher: &stubTokenRefresher{token: "new-token"},
		client:         mockClient,
		control:        ctrl,
		scope:          logtest.NewScope(t),
	}

	_, err := s.applyInstructions(context.Background(), []extension.Instruction{
		{Type: extension.InstructionFetchCredentials},
		{Type: extension.InstructionCloseSyncStream},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if ctrl.notifyTokenRefreshedCalls != 1 {
		t.Fatalf("NotifyTokenRefreshed calls = %d, want 1", ctrl.notifyTokenRefreshedCalls)
	}
}

type stubController struct {
	notifyTokenRefreshedCalls int
	notifyTokenRefreshedErr   error
}

func (s *stubController) Start(context.Context, extension.StartRequest) ([]extension.Instruction, error) {
	return nil, nil
}

func (s *stubController) SendTextLine(context.Context, string) ([]extension.Instruction, error) {
	return nil, nil
}

func (s *stubController) NotifyConnection(context.Context, extension.ConnectionEvent) ([]extension.Instruction, error) {
	return nil, nil
}

func (s *stubController) NotifyTokenRefreshed(context.Context) ([]extension.Instruction, error) {
	s.notifyTokenRefreshedCalls++
	return nil, s.notifyTokenRefreshedErr
}

func (s *stubController) NotifyUploadCompleted(context.Context) ([]extension.Instruction, error) {
	return nil, nil
}

func (s *stubController) Close() error {
	return nil
}

type stubTokenRefresher struct {
	token string
	err   error
}

func (s *stubTokenRefresher) GetAccessToken(context.Context) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.token, nil
}

func (s *stubTokenRefresher) ForceRefreshAccessToken(context.Context) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.token, nil
}
