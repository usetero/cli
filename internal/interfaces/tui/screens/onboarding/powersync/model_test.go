package powersync

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
	sessionruntime "github.com/usetero/cli/internal/runtime/session"
)

type sessionStub struct {
	status sessionruntime.Status
}

func (s sessionStub) Status() sessionruntime.Status {
	return s.status
}

func TestModel_ReadyView(t *testing.T) {
	model := New(sessionStub{
		status: sessionruntime.Status{Sync: &pssyncer.Ready{}},
	}, theme.New(false))
	view := model.View().Content
	if !strings.Contains(view, "PowerSync is ready.") {
		t.Fatalf("expected ready message, got %q", view)
	}
}

func TestModel_ErrorView(t *testing.T) {
	model := New(sessionStub{
		status: sessionruntime.Status{Sync: &pssyncer.Error{Err: errors.New("boom")}},
	}, theme.New(false))
	view := model.View().Content
	if !strings.Contains(view, "Sync failed: boom") {
		t.Fatalf("expected sync failure message, got %q", view)
	}
}

func TestModel_SpinnerTickUpdates(t *testing.T) {
	model := New(sessionStub{
		status: sessionruntime.Status{Sync: &pssyncer.Connecting{}},
	}, theme.New(false))

	msg := model.Init()()
	_, cmd := model.Update(msg)
	if cmd == nil {
		t.Fatal("expected spinner tick command")
	}
}

func TestModel_SyncingProgressView(t *testing.T) {
	model := New(sessionStub{
		status: sessionruntime.Status{
			Sync: &pssyncer.Syncing{Progress: &pssyncer.Progress{Downloaded: 3, Total: 10}},
		},
	}, theme.New(false))
	view := model.View().Content
	if !strings.Contains(view, "3 / 10 rows") {
		t.Fatalf("expected row progress message, got %q", view)
	}
	if !strings.Contains(view, "%") {
		t.Fatalf("expected progress bar output, got %q", view)
	}
}

func TestModel_IgnoreNonTick(t *testing.T) {
	model := New(sessionStub{
		status: sessionruntime.Status{Sync: &pssyncer.Connecting{}},
	}, theme.New(false))
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("expected nil command for non-tick msg")
	}
}
