package datadog

import (
	"context"
	"testing"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/boundary/graphql/apitest"
	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
)

func TestDiscoveryPollTickSchedulesAsyncFetch(t *testing.T) {
	t.Parallel()

	callCount := 0
	mockDatadog := apitest.NewMockDatadogAccounts()
	mockDatadog.GetStatusFunc = func(context.Context, domain.DatadogAccountID) (*graphql.DatadogAccountStatus, error) {
		callCount++
		return &graphql.DatadogAccountStatus{}, nil
	}
	services := apitest.NewMockServiceSet(nil, nil, nil, mockDatadog)

	m := NewDiscovery(context.Background(), styles.NewTheme(true), "dd-1", services, logtest.NewScope(t))
	cmd := m.Update(pollTickMsg{})
	if cmd == nil {
		t.Fatalf("expected poll tick to schedule async fetch")
	}
	if callCount != 0 {
		t.Fatalf("expected no network call until fetch command executes")
	}

	msg := cmd()
	if _, ok := msg.(statusMsg); !ok {
		t.Fatalf("expected statusMsg from async fetch, got %T", msg)
	}
	if callCount != 1 {
		t.Fatalf("expected one network call after fetch command execution, got %d", callCount)
	}
}

func TestDiscoveryStatusSchedulesTimerTick(t *testing.T) {
	t.Parallel()

	m := NewDiscovery(
		context.Background(),
		styles.NewTheme(true),
		"dd-1",
		apitest.NewMockServiceSet(nil, nil, nil, apitest.NewMockDatadogAccounts()),
		logtest.NewScope(t),
	)

	cmd := m.Update(statusMsg{status: &graphql.DatadogAccountStatus{ReadyForUse: false}})
	if cmd == nil {
		t.Fatalf("expected non-ready status to schedule timer")
	}

	msg := cmd()
	if _, ok := msg.(pollTickMsg); !ok {
		t.Fatalf("expected pollTickMsg from scheduled timer, got %T", msg)
	}
}

func TestDiscoveryStatusReadyCompletesStep(t *testing.T) {
	t.Parallel()

	m := NewDiscovery(
		context.Background(),
		styles.NewTheme(true),
		"dd-1",
		apitest.NewMockServiceSet(nil, nil, nil, apitest.NewMockDatadogAccounts()),
		logtest.NewScope(t),
	)

	cmd := m.Update(statusMsg{status: &graphql.DatadogAccountStatus{ReadyForUse: true}})
	if cmd == nil {
		t.Fatalf("expected ready status to emit completion message")
	}
	if _, ok := cmd().(bootstrap.DatadogDiscoveryComplete); !ok {
		t.Fatalf("expected DatadogDiscoveryComplete message")
	}
}
