package chat

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/app/chat/usecase"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
)

// stubStatus implements graphql.Status for the empty-state tests.
type stubStatus struct {
	summary domain.AccountSummary
}

func (s stubStatus) GetAccountSummary(context.Context) (domain.AccountSummary, error) {
	return s.summary, nil
}

func (s stubStatus) ListServiceStatuses(context.Context) ([]domain.ServiceStatus, error) {
	return nil, nil
}

func (s stubStatus) ListServiceLogEvents(context.Context, string) ([]domain.LogEventStatus, error) {
	return nil, nil
}

func newEmptyStateChat(t *testing.T, summary domain.AccountSummary) *Model {
	t.Helper()
	return New(
		nil,
		domain.Account{ID: "acct-1"},
		domain.Workspace{ID: "ws-1"},
		styles.NewTheme(true),
		graphql.ServiceSet{Status: stubStatus{summary: summary}},
		usecase.RuntimeDeps{EffectContext: context.Background()},
		nil,
		logtest.NewScope(t),
	)
}

func TestEmptyStatePollDoesNotMutateSynchronously(t *testing.T) {
	m := newEmptyStateChat(t, domain.AccountSummary{ServiceCount: 7})

	cmd := m.Update(emptyStatePollTickMsg{})
	if cmd == nil {
		t.Fatalf("expected poll update to return command")
	}
	if m.policySummary != nil {
		t.Fatalf("expected summary to remain unset until async result message")
	}
}

func TestEmptyStateSummaryMessageUpdatesState(t *testing.T) {
	m := newEmptyStateChat(t, domain.AccountSummary{ServiceCount: 5, ActiveServices: 3})

	msg := m.fetchEmptyStateSummary()()
	if _, ok := msg.(emptyStateSummaryLoadedMsg); !ok {
		t.Fatalf("expected emptyStateSummaryLoadedMsg, got %T", msg)
	}

	m.Update(msg)
	if m.policySummary == nil {
		t.Fatalf("expected summary after handling async summary message")
	}
	if m.policySummary.ServiceCount != 5 {
		t.Fatalf("unexpected service count: got %d want %d", m.policySummary.ServiceCount, 5)
	}
}
