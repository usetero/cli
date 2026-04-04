package account

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/infrastructure/logging"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/events"
	accountruntime "github.com/usetero/cli/internal/runtime/account"
)

type RuntimeFactory interface {
	New(ctx context.Context, scope accountruntime.Scope) (*accountruntime.Runtime, error)
}

type Model struct {
	scope   logging.Scope
	factory RuntimeFactory
	runtime *accountruntime.Runtime
	status  accountruntime.Status
}

var _ core.Model = (*Model)(nil)

func New(scope logging.Scope, factory RuntimeFactory) *Model {
	if factory == nil {
		panic("account runtime factory is required")
	}
	return &Model{
		scope:   scope,
		factory: factory,
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case events.AccountSelectedMsg:
		return m, m.replace(context.Background(), typed.Scope)
	case RuntimeEventMsg:
		m.refreshStatus()
		return m, tea.Batch(m.waitEvent(), m.publishStatus())
	case RuntimeClosedMsg:
		m.runtime = nil
		m.status = accountruntime.Status{}
		return m, m.publishStatus()
	}
	return m, nil
}

func (m *Model) View() tea.View   { return tea.NewView("") }
func (m *Model) SetSize(_, _ int) {}

func (m *Model) Status() accountruntime.Status {
	return m.status
}

func (m *Model) Close(ctx context.Context) error {
	if m.runtime == nil {
		return nil
	}
	runtime := m.runtime
	m.runtime = nil
	return runtime.Close(ctx)
}

func (m *Model) replace(ctx context.Context, scope accountruntime.Scope) tea.Cmd {
	if err := m.Close(ctx); err != nil {
		m.scope.Error("close prior account runtime", "error", err)
	}

	runtime, err := m.factory.New(ctx, scope)
	if err != nil {
		m.scope.Error("start account runtime", "error", err)
		m.runtime = nil
		m.status = accountruntime.Status{
			Scope: scope,
			Sync:  &pssyncer.Error{Err: err},
		}
		return m.publishStatus()
	}
	m.runtime = runtime
	m.refreshStatus()
	return tea.Batch(m.waitEvent(), m.publishStatus())
}

func (m *Model) waitEvent() tea.Cmd {
	if m.runtime == nil {
		return nil
	}
	runtime := m.runtime
	return func() tea.Msg {
		event, ok := <-runtime.Events()
		if !ok {
			return RuntimeClosedMsg{}
		}
		return RuntimeEventMsg{Event: event}
	}
}

func (m *Model) refreshStatus() {
	if m.runtime == nil {
		m.status = accountruntime.Status{}
		return
	}
	m.status = m.runtime.Status()
}

func (m *Model) publishStatus() tea.Cmd {
	status := m.status
	return func() tea.Msg {
		return events.AccountRuntimeUpdatedMsg{Status: status}
	}
}
