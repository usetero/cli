package datadogselectregion

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
	"github.com/usetero/cli/internal/datadog"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/list"
	"github.com/usetero/cli/internal/tui/keymap"
	datadogapikey "github.com/usetero/cli/internal/tui/onboarding/datadog/apikey"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

// RegionListItem wraps a datadog region for the list.
type RegionListItem struct {
	Site        string
	Domain      string
	DisplayName string
}

func (i RegionListItem) FilterValue() string { return i.DisplayName }

// regionDelegate renders each region in the list.
type regionDelegate struct {
	theme *styles.Theme
}

func (d regionDelegate) Height() int                                    { return 1 }
func (d regionDelegate) Spacing() int                                   { return 0 }
func (d regionDelegate) Update(_ tea.Msg, _ *bubbleslist.Model) tea.Cmd { return nil }
func (d regionDelegate) Render(w io.Writer, m bubbleslist.Model, index int, item bubbleslist.Item) {
	i, ok := item.(RegionListItem)
	if !ok {
		return
	}

	colors := d.theme.Colors

	if index == m.Index() {
		nameStyle := lipgloss.NewStyle().Foreground(colors.Accent).Bold(true)
		domainStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)
		fmt.Fprintf(w, "%s  %s", nameStyle.Render("> "+i.DisplayName), domainStyle.Render(i.Domain))
	} else {
		nameStyle := lipgloss.NewStyle().Foreground(colors.Page.Text)
		domainStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted)
		fmt.Fprintf(w, "%s  %s", nameStyle.Render("  "+i.DisplayName), domainStyle.Render(i.Domain))
	}
}

// Model handles Datadog region selection.
type Model struct {
	ctx   context.Context
	theme *styles.Theme

	role    string
	org     domain.Organization
	account domain.Account

	services api.APIServices
	prefs    preferences.Preferences
	logger   log.Logger

	list           list.Model
	selectedRegion string
	width          int
	height         int
}

// New creates a new region select model.
func New(
	ctx context.Context,
	theme *styles.Theme,
	role string,
	org domain.Organization,
	account domain.Account,
	services api.APIServices,
	prefs preferences.Preferences,
	logger log.Logger,
) Model {
	regions := datadog.GetRegions()

	items := make([]list.Item, len(regions))
	for i, region := range regions {
		items[i] = RegionListItem{
			Site:        region.Site,
			Domain:      region.Domain,
			DisplayName: region.DisplayName,
		}
	}

	delegate := regionDelegate{theme: theme}
	l := list.New(theme, items, delegate)

	return Model{
		ctx:      ctx,
		theme:    theme,
		role:     role,
		org:      org,
		account:  account,
		services: services,
		prefs:    prefs,
		logger:   logger,
		list:     l,
		width:    80,
	}
}

// Init initializes the step.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(RegionListItem); ok {
				m.selectedRegion = item.Site
				m.logger.Info("datadog region selected", "site", item.Site, "name", item.DisplayName)
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the region selection UI.
func (m Model) View() string {
	themeStyles := m.theme.Styles
	colors := m.theme.Colors

	title := themeStyles.Title.Render("Connect to Datadog")
	explanation := themeStyles.Help.Render("Tero needs two keys to access your data - that's how Datadog works.")
	question := themeStyles.Body.Render("Which region is your Datadog account in?")

	linkStyle := lipgloss.NewStyle().Foreground(colors.Page.TextMuted).Underline(true)
	docsLink := themeStyles.Help.Render("Need help? ") + linkStyle.Render("docs.usetero.com/integrations/datadog")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		explanation,
		"",
		question,
		"",
		m.list.View(),
		"",
		docsLink,
	)
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) step.Step {
	m.width = width
	m.height = height
	m.list = m.list.SetSize(width, 10)
	return m
}

// IsBusy returns false.
func (m Model) IsBusy() bool {
	return false
}

// HasError returns false.
func (m Model) HasError() bool {
	return false
}

// Error returns nil.
func (m Model) Error() error {
	return nil
}

// Help returns the key bindings for this step.
func (m Model) Help() help.KeyMap {
	listKeys := m.list.KeyMap()
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
	if m.selectedRegion == "" {
		return nil, step.ErrNotReady
	}

	return datadogapikey.New(
		m.ctx,
		m.theme,
		m.role,
		m.org,
		m.account,
		m.selectedRegion,
		m.services,
		m.prefs,
		m.logger,
	), nil
}

// Close releases resources.
func (m Model) Close() error {
	return nil
}
