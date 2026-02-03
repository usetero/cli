package workspaceselect

import (
	"context"
	"fmt"
	"io"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	bubbleslist "charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/list"
	"github.com/usetero/cli/internal/tui/components/remotelist"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// WorkspaceSelectedMsg is sent when a workspace is selected during onboarding.
// This triggers PowerSync to start syncing in the background.
type WorkspaceSelectedMsg struct {
	Organization domain.Organization
	Account      domain.Account
	Workspace    domain.Workspace
}

// WorkspaceListItem wraps domain.Workspace to implement list.Item.
type WorkspaceListItem struct {
	domain.Workspace
}

func (i WorkspaceListItem) FilterValue() string { return i.Name }

// workspaceDelegate renders each workspace in the list.
type workspaceDelegate struct {
	theme *styles.Theme
}

func (d workspaceDelegate) Height() int                                    { return 1 }
func (d workspaceDelegate) Spacing() int                                   { return 0 }
func (d workspaceDelegate) Update(_ tea.Msg, _ *bubbleslist.Model) tea.Cmd { return nil }
func (d workspaceDelegate) Render(w io.Writer, m bubbleslist.Model, index int, item bubbleslist.Item) {
	i, ok := item.(WorkspaceListItem)
	if !ok {
		return
	}

	colors := d.theme.Colors
	str := i.Name
	if index == m.Index() {
		fmt.Fprint(w, lipgloss.NewStyle().Foreground(colors.Accent).Bold(true).Render("> "+str))
	} else {
		fmt.Fprint(w, lipgloss.NewStyle().Foreground(colors.Page.Text).Render("  "+str))
	}
}

// Model handles selecting a workspace.
type Model struct {
	ctx     context.Context
	theme   *styles.Theme
	org     domain.Organization
	account domain.Account

	services api.APIServices
	prefs    preferences.Preferences
	logger   log.Logger

	remoteList          remotelist.Model
	workspaceList       []domain.Workspace
	selectedWorkspaceID domain.WorkspaceID
	selectedWorkspace   domain.Workspace
	width               int
	height              int
}

// New creates a new workspace select model.
func New(
	ctx context.Context,
	theme *styles.Theme,
	org domain.Organization,
	account domain.Account,
	services api.APIServices,
	prefs preferences.Preferences,
	logger log.Logger,
) Model {
	delegate := workspaceDelegate{theme: theme}

	return Model{
		ctx:        ctx,
		theme:      theme,
		org:        org,
		account:    account,
		services:   services,
		prefs:      prefs,
		logger:     logger,
		remoteList: remotelist.New(theme, delegate, "Loading workspaces", logger),
		width:      80,
	}
}

// Init starts loading workspaces.
func (m Model) Init() tea.Cmd {
	_, cmd := m.remoteList.InitWithLoader(m.loadWorkspaces())
	return cmd
}

// loadWorkspaces returns a command that loads workspaces.
func (m Model) loadWorkspaces() tea.Cmd {
	return func() tea.Msg {
		m.logger.Info("loading workspaces", "accountID", m.account.ID)

		workspaces, err := m.services.Workspaces.List(m.ctx, m.account.ID.String())
		if err != nil {
			m.logger.Error("failed to load workspaces", "error", err)
			return remotelist.LoadResultMsg{Items: nil, Err: err}
		}

		m.logger.Info("workspaces loaded", "count", len(workspaces))

		items := make([]list.Item, len(workspaces))
		for i, ws := range workspaces {
			items[i] = WorkspaceListItem{Workspace: ws}
		}

		return remotelist.LoadResultMsg{Items: items, Err: nil}
	}
}

// selectWorkspace handles workspace selection.
func (m Model) selectWorkspace(workspace domain.Workspace) Model {
	m.selectedWorkspaceID = workspace.ID
	m.selectedWorkspace = workspace
	return m
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case remotelist.LoadResultMsg:
		if msg.Err == nil {
			m.workspaceList = make([]domain.Workspace, 0, len(msg.Items))
			for _, item := range msg.Items {
				if wsItem, ok := item.(WorkspaceListItem); ok {
					m.workspaceList = append(m.workspaceList, wsItem.Workspace)
				}
			}

			userPref := m.prefs.GetDefaultWorkspaceID()

			// Has preference and exists → auto-select
			if userPref != "" {
				for _, ws := range m.workspaceList {
					if ws.ID == userPref {
						m.logger.Debug("auto-selected workspace from preference", "id", userPref, "name", ws.Name)
						m = m.selectWorkspace(ws)
						// Emit WorkspaceSelectedMsg to trigger sync
						selectedWs := ws
						cmds = append(cmds, func() tea.Msg {
							return WorkspaceSelectedMsg{Organization: m.org, Account: m.account, Workspace: selectedWs}
						})
						break
					}
				}
			}

			// No preference and only 1 workspace → auto-select and save
			if userPref == "" && len(m.workspaceList) == 1 {
				ws := m.workspaceList[0]
				m.logger.Info("auto-selected workspace", "id", ws.ID, "name", ws.Name, "reason", "only one available")
				if err := m.prefs.SetDefaultWorkspaceID(ws.ID); err != nil {
					m.logger.Error("failed to save workspace preference", "error", err)
				}
				m.selectedWorkspaceID = ws.ID
				m.selectedWorkspace = ws
				// Emit WorkspaceSelectedMsg to trigger sync
				cmds = append(cmds, func() tea.Msg {
					return WorkspaceSelectedMsg{Organization: m.org, Account: m.account, Workspace: ws}
				})
			}
		}

	case tea.KeyPressMsg:
		switch msg.String() {
		case "r":
			if m.remoteList.HasError() {
				m.logger.Info("user requested retry")
				var cmd tea.Cmd
				m.remoteList, cmd = m.remoteList.Retry()
				cmds = append(cmds, cmd)
			}
		case "enter":
			if !m.remoteList.IsLoaded() {
				break
			}
			selected := m.remoteList.SelectedItem()
			if wsItem, ok := selected.(WorkspaceListItem); ok {
				m.logger.Info("workspace selected", "id", wsItem.ID, "name", wsItem.Name)
				if err := m.prefs.SetDefaultWorkspaceID(wsItem.ID); err != nil {
					m.logger.Error("failed to save workspace preference", "error", err)
				}
				m = m.selectWorkspace(wsItem.Workspace)
				// Emit WorkspaceSelectedMsg to trigger sync
				ws := wsItem.Workspace
				cmds = append(cmds, func() tea.Msg {
					return WorkspaceSelectedMsg{Organization: m.org, Account: m.account, Workspace: ws}
				})
			}
		}
	}

	var cmd tea.Cmd
	m.remoteList, cmd = m.remoteList.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the workspace selection UI.
func (m Model) View() string {
	themeStyles := m.theme.Styles

	if m.remoteList.IsBusy() {
		return m.remoteList.View()
	}

	if m.remoteList.HasError() {
		return m.remoteList.View()
	}

	title := themeStyles.Title.Render("Select your workspace")
	subtitle := themeStyles.Help.Render("Workspaces organize your conversations and telemetry analysis")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		m.remoteList.View(),
	)
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) step.Step {
	m.width = width
	m.height = height
	m.remoteList = m.remoteList.SetWidth(width)
	return m
}

// IsBusy returns true while loading.
func (m Model) IsBusy() bool {
	return !m.remoteList.IsLoaded()
}

// HasError returns true if there's an error.
func (m Model) HasError() bool {
	return m.remoteList.HasError()
}

// Error returns the current error.
func (m Model) Error() error {
	return m.remoteList.Error()
}

// Help returns the key bindings for this step.
func (m Model) Help() help.KeyMap {
	if m.remoteList.IsBusy() {
		return keymap.Simple{Keys: []key.Binding{}}
	}

	if m.remoteList.HasError() {
		return keymap.Simple{
			Keys: []key.Binding{
				key.NewBinding(
					key.WithKeys("r"),
					key.WithHelp("r", "retry"),
				),
			},
		}
	}

	listKeys := m.remoteList.KeyMap()
	return keymap.Simple{
		Keys: []key.Binding{
			listKeys.CursorUp,
			listKeys.CursorDown,
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select"),
			),
		},
	}
}

// Next returns the next step.
func (m Model) Next() (step.Step, error) {
	if m.selectedWorkspaceID == "" {
		return nil, step.ErrNotReady
	}

	// TODO: return sync step
	return nil, nil
}

// Organization returns the organization.
func (m Model) Organization() domain.Organization {
	return m.org
}

// Account returns the account.
func (m Model) Account() domain.Account {
	return m.account
}

// Workspace returns the selected workspace.
func (m Model) Workspace() domain.Workspace {
	return m.selectedWorkspace
}

// Close releases resources.
func (m Model) Close() error {
	return nil
}
