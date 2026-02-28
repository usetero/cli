package chat

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
	"github.com/usetero/cli/internal/styles"
)

type stubDatadogAccountStatuses struct {
	getSummary func(ctx context.Context) (domain.AccountSummary, error)
}

func (s stubDatadogAccountStatuses) GetSummary(ctx context.Context) (domain.AccountSummary, error) {
	return s.getSummary(ctx)
}

func TestEmptyStatePollDoesNotMutateSynchronously(t *testing.T) {
	mockDB := sqlitetest.NewMockDB()
	mockDB.DatadogAccountStatusesImpl = stubDatadogAccountStatuses{
		getSummary: func(context.Context) (domain.AccountSummary, error) {
			return domain.AccountSummary{ServiceCount: 7}, nil
		},
	}

	m := New(
		nil,
		domain.Account{ID: "acct-1"},
		domain.Workspace{ID: "ws-1"},
		styles.NewTheme(true),
		mockDB,
		nil,
		nil,
		logtest.NewScope(t),
	)

	cmd := m.Update(emptyStatePollMsg{})
	if cmd == nil {
		t.Fatalf("expected poll update to return command")
	}
	if m.policySummary != nil {
		t.Fatalf("expected summary to remain unset until async result message")
	}
}

func TestEmptyStateSummaryMessageUpdatesState(t *testing.T) {
	mockDB := sqlitetest.NewMockDB()
	mockDB.DatadogAccountStatusesImpl = stubDatadogAccountStatuses{
		getSummary: func(context.Context) (domain.AccountSummary, error) {
			return domain.AccountSummary{ServiceCount: 5, ActiveServices: 3}, nil
		},
	}

	m := New(
		nil,
		domain.Account{ID: "acct-1"},
		domain.Workspace{ID: "ws-1"},
		styles.NewTheme(true),
		mockDB,
		nil,
		nil,
		logtest.NewScope(t),
	)

	msg := m.fetchEmptyStateSummary()()
	if _, ok := msg.(emptyStateSummaryMsg); !ok {
		t.Fatalf("expected emptyStateSummaryMsg, got %T", msg)
	}

	m.Update(msg)
	if m.policySummary == nil {
		t.Fatalf("expected summary after handling async summary message")
	}
	if m.policySummary.ServiceCount != 5 {
		t.Fatalf("unexpected service count: got %d want %d", m.policySummary.ServiceCount, 5)
	}
}
