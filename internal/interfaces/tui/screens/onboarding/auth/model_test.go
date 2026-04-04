package auth

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	identity "github.com/usetero/cli/internal/domains/identity"
	"github.com/usetero/cli/internal/domains/identity/identitytest"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

func TestAuthInitStartsWaitingFlow(t *testing.T) {
	t.Parallel()

	provider := identitytest.NewProvider()
	provider.StartDeviceAuthFn = func(context.Context) (identity.DeviceFlow, error) {
		return identity.DeviceFlow{
			UserCode:                "ABCD-EFGH",
			VerificationURIComplete: "https://example.com/verify",
			DeviceCode:              "device_1",
			Interval:                time.Millisecond,
		}, nil
	}
	provider.PollAuthenticationFn = func(context.Context, string) (identity.Tokens, identity.User, error) {
		return identity.Tokens{
			AccessToken:  "a1",
			RefreshToken: "r1",
		}, identity.User{ID: "user_1", Email: "u@example.com"}, nil
	}

	service := identity.NewService(provider, identitytest.NewTokenStore(), identity.NopLogger{})
	model := New(logging.Scope{}, service, theme.New(false))

	opened := ""
	previousOpenBrowser := openBrowser
	openBrowser = func(url string) error {
		opened = url
		return nil
	}
	defer func() { openBrowser = previousOpenBrowser }()

	if cmd := model.Init(); cmd != nil {
		t.Fatalf("expected auth init to be idle")
	}
	_, cmd := model.Update(keyPress("enter"))
	if model.Busy() == nil {
		t.Fatalf("expected model to become busy after start")
	}
	if model.Input().Label == "" || !strings.Contains(model.Input().Label, "Welcome to Tero") {
		t.Fatalf("expected auth label, got %#v", model.Input())
	}

	if cmd == nil {
		t.Fatalf("expected device-flow start command")
	}

	msg := cmd()
	_, cmd = model.Update(msg)
	if model.Busy() == nil {
		t.Fatalf("expected model to remain busy after device flow start")
	}

	browserMsg := model.openBrowser()()
	if typed, ok := browserMsg.(browserOpenedMsg); !ok || typed.Err != nil {
		t.Fatalf("expected browser open success, got %#v", browserMsg)
	}

	if opened != "https://example.com/verify" {
		t.Fatalf("expected browser URL to be opened, got %q", opened)
	}

	pollMsg := model.pollDeviceFlow()()
	if _, ok := pollMsg.(deviceFlowCompletedMsg); !ok {
		t.Fatalf("expected auth completion message, got %#v", pollMsg)
	}
	_, _ = model.Update(pollMsg)
	if model.Busy() != nil {
		t.Fatalf("expected model to stop being busy after auth completion")
	}
}

func TestAuthViewShowsFailureState(t *testing.T) {
	t.Parallel()

	service := identity.NewService(identitytest.NewProvider(), identitytest.NewTokenStore(), identity.NopLogger{})
	model := New(logging.Scope{}, service, theme.New(false))

	_, _ = model.Update(deviceFlowFailedMsg{Err: context.DeadlineExceeded})

	input := model.Input()
	if input == nil || !strings.Contains(strings.ToLower(input.Action), "try again") {
		t.Fatalf("expected retry action, got %#v", input)
	}
}

func TestAuthWaitingHelpIncludesReopenBinding(t *testing.T) {
	t.Parallel()

	service := identity.NewService(identitytest.NewProvider(), identitytest.NewTokenStore(), identity.NopLogger{})
	model := New(logging.Scope{}, service, theme.New(false))

	_, _ = model.Update(deviceFlowStartedMsg{
		Flow: identity.DeviceFlow{
			UserCode:                "CODE",
			VerificationURIComplete: "https://example.com/verify",
			DeviceCode:              "device_1",
			Interval:                time.Second,
		},
	})

	bindings := model.ShortHelp()
	if len(bindings) != 1 || bindings[0].Help().Key == "" {
		t.Fatalf("expected reopen browser binding, got %+v", bindings)
	}
}

func TestAuthIdleStartsWithActionInput(t *testing.T) {
	t.Parallel()

	service := identity.NewService(identitytest.NewProvider(), identitytest.NewTokenStore(), identity.NopLogger{})
	model := New(logging.Scope{}, service, theme.New(false))

	input := model.Input()
	if input == nil || input.Kind != core.InputConfirm {
		t.Fatalf("expected idle action input, got %#v", input)
	}
	if !strings.Contains(strings.ToLower(input.Action), "get started") {
		t.Fatalf("expected start action, got %#v", input)
	}

	bindings := model.ShortHelp()
	if len(bindings) != 0 {
		t.Fatalf("expected no page-owned start binding, got %+v", bindings)
	}
}

func keyPress(key string) tea.KeyPressMsg {
	switch key {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	default:
		return tea.KeyPressMsg{Text: key}
	}
}
