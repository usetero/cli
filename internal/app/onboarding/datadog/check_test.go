package datadog

import (
	"context"
	"errors"
	"testing"

	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
)

func TestCheckHasDatadogEmitsReady(t *testing.T) {
	t.Parallel()

	mockDatadog := apitest.NewMockDatadogAccounts()
	mockDatadog.HasAccountFunc = func(context.Context, domain.AccountID) (bool, error) {
		return true, nil
	}
	services := apitest.NewMockAPIServices(nil, nil, nil, mockDatadog)

	m := NewCheck(context.Background(), styles.NewTheme(true), domain.Account{ID: "acc-1"}, services, logtest.NewScope(t))
	cmd := m.Update(checkResultMsg{hasDatadog: true})
	if cmd == nil {
		t.Fatal("expected command")
	}
	if _, ok := cmd().(bootstrap.DatadogReady); !ok {
		t.Fatalf("expected DatadogReady message")
	}
}

func TestCheckNoDatadogEmitsNeeded(t *testing.T) {
	t.Parallel()

	mockDatadog := apitest.NewMockDatadogAccounts()
	mockDatadog.HasAccountFunc = func(context.Context, domain.AccountID) (bool, error) {
		return false, nil
	}
	services := apitest.NewMockAPIServices(nil, nil, nil, mockDatadog)

	m := NewCheck(context.Background(), styles.NewTheme(true), domain.Account{ID: "acc-1"}, services, logtest.NewScope(t))
	cmd := m.Update(checkResultMsg{hasDatadog: false})
	if cmd == nil {
		t.Fatal("expected command")
	}
	if _, ok := cmd().(bootstrap.DatadogNeeded); !ok {
		t.Fatalf("expected DatadogNeeded message")
	}
}

func TestCheckErrorEnablesRetry(t *testing.T) {
	t.Parallel()

	mockDatadog := apitest.NewMockDatadogAccounts()
	mockDatadog.HasAccountFunc = func(context.Context, domain.AccountID) (bool, error) {
		return false, errors.New("boom")
	}
	services := apitest.NewMockAPIServices(nil, nil, nil, mockDatadog)

	m := NewCheck(context.Background(), styles.NewTheme(true), domain.Account{ID: "acc-1"}, services, logtest.NewScope(t))
	if cmd := m.Update(checkResultMsg{err: errors.New("boom")}); cmd == nil {
		t.Fatal("expected error command")
	}
	if len(m.ShortHelp()) == 0 {
		t.Fatal("expected retry keybinding after error")
	}
}
