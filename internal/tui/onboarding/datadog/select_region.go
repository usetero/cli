package datadog

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/pkg/browser"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/datadog"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/components/list"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/styles"
)

// regionItem implements list.Item for the list component
type regionItem struct {
	site        string
	domain      string
	displayName string
}

func (i regionItem) FilterValue() string { return i.displayName }

// regionDelegate renders each region in the list
type regionDelegate struct{}

func (d regionDelegate) Height() int                             { return 1 }
func (d regionDelegate) Spacing() int                            { return 0 }
func (d regionDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d regionDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(regionItem)
	if !ok {
		return
	}

	theme := styles.CurrentTheme()

	if index == m.Index() {
		nameStyle := lipgloss.NewStyle().
			Foreground(theme.Primary).
			Bold(true)
		domainStyle := lipgloss.NewStyle().
			Foreground(theme.TextSubtle)
		_, _ = fmt.Fprintf(w, "%s  %s", nameStyle.Render("> "+i.displayName), domainStyle.Render(i.domain))
	} else {
		nameStyle := lipgloss.NewStyle().
			Foreground(theme.Text)
		domainStyle := lipgloss.NewStyle().
			Foreground(theme.TextMuted)
		_, _ = fmt.Fprintf(w, "%s  %s", nameStyle.Render("  "+i.displayName), domainStyle.Render(i.domain))
	}
}

// SelectRegionStep handles greeting and Datadog region selection
type SelectRegionStep struct {
	// Accumulated state from previous steps
	role      string
	orgID     string
	accountID string

	// Pass-through to next step
	apiClient api.Client
	logger    log.Logger

	// UI state
	list           *list.List
	selectedRegion string
	width          int
	globalBindings []key.Binding
}

// NewSelectRegionStep creates a new Datadog region selection step
func NewSelectRegionStep(role string, orgID string, accountID string, apiClient api.Client, logger log.Logger, globalBindings []key.Binding) step.Step {
	if apiClient == nil {
		panic("apiClient cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	regions := datadog.GetRegions()

	// Convert to list items
	items := make([]list.Item, len(regions))
	for i, region := range regions {
		items[i] = regionItem{
			site:        region.Site,
			domain:      region.Domain,
			displayName: region.DisplayName,
		}
	}

	delegate := regionDelegate{}
	l := list.New(items, delegate)

	return &SelectRegionStep{
		role:           role,
		orgID:          orgID,
		accountID:      accountID,
		apiClient:      apiClient,
		logger:         logger,
		list:           l,
		width:          80,
		globalBindings: globalBindings,
	}
}

// Init initializes the step
func (s *SelectRegionStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *SelectRegionStep) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "o":
			// Open Datadog integration docs
			url := "https://usetero.com/docs/integrations/datadog"
			err := browser.OpenURL(url)
			if err != nil {
				s.logger.Error("failed to open browser", "error", err, "url", url)
			} else {
				s.logger.Debug("opened browser for datadog docs", "url", url)
			}
			return s, nil
		case "enter":
			if item, ok := s.list.SelectedItem().(regionItem); ok {
				s.selectedRegion = item.site
				s.logger.Info("datadog region selected", log.String("site", item.site), log.String("name", item.displayName))
			}
			return s, nil
		}
	}

	cmd := s.list.Update(msg)
	return s, cmd
}

// View renders the region selection UI
func (s *SelectRegionStep) View() string {
<<<<<<< HEAD
	theme := styles.CurrentTheme()

	header := RenderHeader()

	stepTitleStyle := lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true)

	stepTitle := stepTitleStyle.Render("Step 1 of 3: Select your region")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(theme.TextMuted)
	subtitle := subtitleStyle.Render("Select the region that matches your Datadog URL:")
=======
	common := styles.Common()

	header := RenderHeader()

	stepTitle := common.Title.Render("Step 1 of 3: Select your region")
	subtitle := common.Help.Render("Select the region that matches your Datadog URL:")
>>>>>>> 17e8dd9 (chore: initial commit)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		stepTitle,
		"",
		subtitle,
		"",
		s.list.View(),
	)
}

// SetSize sets the width available for rendering
func (s *SelectRegionStep) SetSize(width, height int) {
	s.width = width
	s.list.SetSize(width, 10)
}

// IsComplete returns true if a region has been selected
func (s *SelectRegionStep) IsComplete() bool {
	return s.selectedRegion != ""
}

// SelectedRegion returns the selected Datadog region site identifier
func (s *SelectRegionStep) SelectedRegion() string {
	return s.selectedRegion
}

// IsBusy returns false - region selection is never busy
func (s *SelectRegionStep) IsBusy() bool {
	return false
}

// HasError returns false - this step has no error state
func (s *SelectRegionStep) HasError() bool {
	return false
}

// Error returns nil - this step has no error state
func (s *SelectRegionStep) Error() error {
	return nil
}

// Next returns the next step after Datadog region selection
func (s *SelectRegionStep) Next() step.Step {
	// Create Datadog account service for next step
	datadogAccountService := api.NewDatadogAccountService(s.apiClient, s.logger)

	// Region selected, continue to API key entry with the selected site
	return NewAPIKeyStep(s.role, s.orgID, s.accountID, s.selectedRegion, datadogAccountService, s.apiClient, s.logger, s.globalBindings)
}

// Help returns the key bindings for this step
func (s *SelectRegionStep) Help() help.KeyMap {
	// Get the list's KeyMap and expose only the bindings we use
	listKeys := s.list.KeyMap()
	return keymap.Simple{
		Keys: []key.Binding{
			listKeys.CursorUp,
			listKeys.CursorDown,
			key.NewBinding(
				key.WithKeys("o"),
				key.WithHelp("o", "learn more"),
			),
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select"),
			),
		},
	}
}
