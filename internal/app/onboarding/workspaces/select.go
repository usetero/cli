// Package workspaces provides workspace selection steps.
package workspaces

import (
	"context"
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/remotelist"
)

// SelectModel handles workspace selection.
type SelectModel struct {
	ctx      context.Context
	theme    *styles.Theme
	services api.APIServices
	prefs    preferences.Preferences
	account  domain.Account
	logger   log.Logger

	list       *remotelist.Model
	workspaces []domain.Workspace
	width      int
	height     int
}

// NewSelect creates a new workspace select step.
func NewSelect(
	ctx context.Context,
	theme *styles.Theme,
	account domain.Account,
	services api.APIServices,
	prefs preferences.Preferences,
	logger log.Logger,
) *SelectModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if theme == nil {
		panic("theme is nil")
	}
	if prefs == nil {
		panic("prefs is nil")
	}
	if logger == nil {
		panic("logger is nil")
	}

	logger = logger.With(slog.String("step", "workspace_select"))
	logger.Debug("initialized")

	return &SelectModel{
		ctx:      ctx,
		theme:    theme,
		services: services,
		prefs:    prefs,
		account:  account,
		logger:   logger,
		list:     remotelist.New(theme, "Loading workspaces"),
	}
}

// Init starts loading workspaces.
func (m *SelectModel) Init() tea.Cmd {
	m.logger.Debug("loading workspaces", slog.String("account_id", m.account.ID.String()))
	return m.list.InitWithLoader(m.loadWorkspaces())
}

func (m *SelectModel) loadWorkspaces() tea.Cmd {
	return func() tea.Msg {
		workspaces, err := m.services.Workspaces.List(m.ctx, m.account.ID.String())
		if err != nil {
			m.logger.Error("failed to load workspaces", slog.Any("error", err))
			return remotelist.LoadResult{Err: err}
		}

		m.logger.Debug("loaded workspaces", slog.Int("count", len(workspaces)))
		items := make([]remotelist.Item, len(workspaces))
		for i, ws := range workspaces {
			items[i] = ws // domain.Workspace implements FilterValue()
		}
		return remotelist.LoadResult{Items: items}
	}
}

// Update handles messages.
func (m *SelectModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case remotelist.LoadResult:
		if msg.Err == nil {
			m.workspaces = make([]domain.Workspace, len(msg.Items))
			for i, item := range msg.Items {
				m.workspaces[i] = item.(domain.Workspace)
			}

			// Check for saved preference
			prefID := m.prefs.GetDefaultWorkspaceID()
			if prefID != "" {
				for _, ws := range m.workspaces {
					if ws.ID == prefID {
						m.logger.Info("workspace restored from preference", slog.String("workspace_id", string(ws.ID)))
						return m.emitSelected(ws)
					}
				}
			}

			// Auto-select if only one
			if len(m.workspaces) == 1 {
				ws := m.workspaces[0]
				_ = m.prefs.SetDefaultWorkspaceID(ws.ID)
				m.logger.Info("workspace auto-selected", slog.String("workspace_id", string(ws.ID)))
				return m.emitSelected(ws)
			}
		}
		return m.list.Update(msg)

	case tea.KeyPressMsg:
		if m.list.IsLoading() {
			return nil
		}
		switch msg.String() {
		case "enter":
			if item := m.list.SelectedItem(); item != nil {
				ws := item.(domain.Workspace)
				_ = m.prefs.SetDefaultWorkspaceID(ws.ID)
				m.logger.Info("workspace selected", slog.String("workspace_id", string(ws.ID)))
				return m.emitSelected(ws)
			}
		case "r":
			if m.list.HasError() {
				m.logger.Debug("retrying workspace load")
				return m.list.Retry()
			}
		}
	}

	return m.list.Update(msg)
}

func (m *SelectModel) emitSelected(ws domain.Workspace) tea.Cmd {
	return func() tea.Msg {
		return msgs.WorkspaceSelected{Workspace: ws}
	}
}

// View renders the workspace selection UI.
func (m *SelectModel) View() string {
	s := m.theme.Styles

	if m.list.IsLoading() {
		return m.list.View()
	}

	if m.list.HasError() {
		return s.Error.Render("Failed to load workspaces. Press 'r' to retry.")
	}

	title := s.Title.Render("Select your workspace")
	subtitle := s.Help.Render("Workspaces organize your conversations")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		m.list.View(),
	)
}

// SetSize updates dimensions.
func (m *SelectModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetWidth(width)
}
