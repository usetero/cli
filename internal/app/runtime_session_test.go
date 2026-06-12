package app

import (
	"context"
	"testing"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/boundary/graphql/apitest"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
)

func TestShutdown_CancelsSession(t *testing.T) {
	cancelled := false
	m := &Model{
		scope:         logtest.NewScope(t),
		sessionCancel: func() { cancelled = true },
	}

	m.shutdown()

	if !cancelled {
		t.Fatalf("expected session cancel to be called")
	}
	if m.sessionCancel != nil {
		t.Fatalf("expected sessionCancel to be cleared")
	}
}

func TestShutdown_NoSessionIsSafe(t *testing.T) {
	m := &Model{scope: logtest.NewScope(t)}
	m.shutdown() // must not panic with no active session
}

func TestStartSession_ScopesServicesToAccount(t *testing.T) {
	scope := logtest.NewScope(t)

	mockClient := apitest.NewMockClient()
	var scopedAccountID domain.AccountID
	mockClient.SetAccountIDFunc = func(accountID domain.AccountID) {
		scopedAccountID = accountID
	}

	m := &Model{
		ctx:      context.Background(),
		scope:    scope,
		services: graphql.NewServiceSetFromClient(mockClient, scope),
	}

	m.startSession("acc_123")
	t.Cleanup(m.shutdown)

	if m.sessionCancel == nil {
		t.Fatalf("expected session cancel to be initialized")
	}
	if m.sessionCtx == nil {
		t.Fatalf("expected session context to be initialized")
	}
	if scopedAccountID != domain.AccountID("acc_123") {
		t.Fatalf("expected services account scope to be set, got %q", scopedAccountID)
	}
}

func TestStartSession_ReplacesPreviousSession(t *testing.T) {
	scope := logtest.NewScope(t)
	m := &Model{
		ctx:      context.Background(),
		scope:    scope,
		services: graphql.NewServiceSetFromClient(apitest.NewMockClient(), scope),
	}

	m.startSession("acc_1")
	firstCtx := m.sessionCtx
	m.startSession("acc_2")
	t.Cleanup(m.shutdown)

	if firstCtx.Err() == nil {
		t.Fatalf("expected previous session context to be cancelled")
	}
	if m.sessionCtx == firstCtx {
		t.Fatalf("expected a new session context")
	}
}
