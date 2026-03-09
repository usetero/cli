package integrationsflow

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/interfaces/tui/components/loading"
	"github.com/usetero/cli/internal/interfaces/tui/present"
	"github.com/usetero/cli/internal/interfaces/tui/screen"
	datadogapikey "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations/datadog/api_key"
	datadogappkey "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations/datadog/app_key"
	datadogregion "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations/datadog/region"
	providerselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/integrations/provider/select"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
	"github.com/usetero/cli/internal/runtime/onboarding"
)

type route int

const (
	routeNone route = iota
	routeProviderSelect
	routeDatadogRegion
	routeDatadogAPIKey
	routeDatadogAppKey
	routeDatadogDiscovery
)

type childID string

const (
	childProviderSelect childID = "provider_select"
	childDatadogRegion  childID = "datadog_region"
	childDatadogAPIKey  childID = "datadog_api_key"
	childDatadogAppKey  childID = "datadog_app_key"
)

var (
	discoveryRefreshEnterBinding = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "refresh"),
	)
	discoveryRefreshRBinding = key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	)
)

type Model struct {
	route  route
	theme  theme.Theme
	router screen.Router[childID]

	providers      []integrations.Provider
	providerSelect *providerselect.Model
	datadogRegion  *datadogregion.Model
	datadogAPIKey  *datadogapikey.Model
	datadogAppKey  *datadogappkey.Model
	discovery      *loading.Model
}

var _ screen.Model = (*Model)(nil)

func New(
	providerSelectModel *providerselect.Model,
	datadogRegionModel *datadogregion.Model,
	datadogAPIKeyModel *datadogapikey.Model,
	datadogAppKeyModel *datadogappkey.Model,
	providers []integrations.Provider,
	appTheme theme.Theme,
) *Model {
	switch {
	case providerSelectModel == nil:
		panic("integrations provider select model is required")
	case datadogRegionModel == nil:
		panic("integrations datadog region model is required")
	case datadogAPIKeyModel == nil:
		panic("integrations datadog api key model is required")
	case datadogAppKeyModel == nil:
		panic("integrations datadog app key model is required")
	}
	model := &Model{
		route:          routeNone,
		theme:          appTheme,
		providers:      append([]integrations.Provider(nil), providers...),
		providerSelect: providerSelectModel,
		datadogRegion:  datadogRegionModel,
		datadogAPIKey:  datadogAPIKeyModel,
		datadogAppKey:  datadogAppKeyModel,
		discovery:      loading.NewThinking(appTheme, "Waiting for Datadog discovery"),
	}
	model.discovery.SetDetail("Press enter or r to refresh.")
	model.router.Register(childProviderSelect, model.providerSelect)
	model.router.Register(childDatadogRegion, model.datadogRegion)
	model.router.Register(childDatadogAPIKey, model.datadogAPIKey)
	model.router.Register(childDatadogAppKey, model.datadogAppKey)

	model.router.SetLift(childProviderSelect, liftProviderSelectCmd)
	model.router.SetLift(childDatadogRegion, liftDatadogRegionCmd)
	model.router.SetLift(childDatadogAPIKey, liftDatadogAPIKeyCmd)
	model.router.SetLift(childDatadogAppKey, liftDatadogAppKeyCmd)
	return model
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) SetSize(width, height int) { m.router.SetSizeAll(width, height) }

func (m *Model) ApplyState(state onboarding.State) bool {
	switch state.NextStep {
	case onboarding.StepDatadogRegion:
		m.datadogAPIKey.Reset()
		m.datadogAppKey.Reset()
		if len(m.providers) <= 1 {
			m.route = routeDatadogRegion
			m.router.ActivateOnly(childDatadogRegion)
			return true
		}
		m.providerSelect.SetProviders(m.providers)
		m.route = routeProviderSelect
		m.router.ActivateOnly(childProviderSelect)
		return true
	case onboarding.StepDatadogAPIKey:
		m.datadogAPIKey.SetSite(state.DatadogDraft.Site)
		m.route = routeDatadogAPIKey
		m.router.ActivateOnly(childDatadogAPIKey)
		return true
	case onboarding.StepDatadogAppKey:
		m.datadogAppKey.SetSite(state.DatadogDraft.Site)
		m.route = routeDatadogAppKey
		m.router.ActivateOnly(childDatadogAppKey)
		return true
	case onboarding.StepDatadogDiscovery:
		m.route = routeDatadogDiscovery
		m.router.ClearActive()
		return true
	default:
		m.route = routeNone
		m.router.ClearActive()
		return false
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var discoveryCmd tea.Cmd
	if m.route == routeDatadogDiscovery {
		next, cmd := m.discovery.Update(msg)
		if model, ok := next.(*loading.Model); ok {
			m.discovery = model
		}
		discoveryCmd = cmd
	}

	switch typed := msg.(type) {
	case ProviderSelectedMsg:
		if typed.Provider == integrations.ProviderDatadog {
			m.route = routeDatadogRegion
			m.router.ActivateOnly(childDatadogRegion)
		}
		return m, discoveryCmd
	}

	switch m.route {
	case routeProviderSelect:
		return m, batch(discoveryCmd, m.router.Forward(msg))
	case routeDatadogRegion, routeDatadogAPIKey, routeDatadogAppKey:
		return m, batch(discoveryCmd, m.router.Forward(msg))
	case routeDatadogDiscovery:
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			if key.Matches(keyMsg, discoveryRefreshEnterBinding) || key.Matches(keyMsg, discoveryRefreshRBinding) {
				return m, batch(discoveryCmd, func() tea.Msg { return RefreshRequestedMsg{} })
			}
		}
		return m, discoveryCmd
	default:
		return m, discoveryCmd
	}
}

func (m *Model) View() tea.View {
	switch m.route {
	case routeProviderSelect:
		return m.providerSelect.View()
	case routeDatadogRegion:
		return m.datadogRegion.View()
	case routeDatadogAPIKey:
		return m.datadogAPIKey.View()
	case routeDatadogAppKey:
		return m.datadogAppKey.View()
	case routeDatadogDiscovery:
		return m.discovery.View()
	default:
		return present.View(m.theme, present.Notice("Integrations flow is not active.", ""))
	}
}

// ShortHelp returns active integration flow key bindings.
func (m *Model) ShortHelp() []key.Binding {
	switch m.route {
	case routeProviderSelect, routeDatadogRegion, routeDatadogAPIKey, routeDatadogAppKey:
		return m.router.ShortHelp()
	case routeDatadogDiscovery:
		return []key.Binding{
			discoveryRefreshEnterBinding,
			discoveryRefreshRBinding,
		}
	default:
		return nil
	}
}

func liftProviderSelectCmd(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		msg := cmd()
		switch typed := msg.(type) {
		case providerselect.SelectedMsg:
			return ProviderSelectedMsg{Provider: typed.Provider}
		default:
			return msg
		}
	}
}

func liftDatadogRegionCmd(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		msg := cmd()
		switch typed := msg.(type) {
		case datadogregion.SelectedMsg:
			return SetDatadogSiteMsg{Site: typed.Site}
		default:
			return msg
		}
	}
}

func liftDatadogAPIKeyCmd(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		msg := cmd()
		switch typed := msg.(type) {
		case datadogapikey.SubmittedMsg:
			return SubmitDatadogAPIKeyMsg{Submission: integrations.DatadogAPIKeySubmission{APIKey: typed.APIKey}}
		default:
			return msg
		}
	}
}

func liftDatadogAppKeyCmd(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		msg := cmd()
		switch typed := msg.(type) {
		case datadogappkey.SubmittedMsg:
			return SubmitDatadogAppKeyMsg{
				Submission: integrations.DatadogAppKeySubmission{
					Name:   typed.Name,
					AppKey: typed.AppKey,
				},
			}
		default:
			return msg
		}
	}
}

func batch(cmds ...tea.Cmd) tea.Cmd {
	filtered := make([]tea.Cmd, 0, len(cmds))
	for _, cmd := range cmds {
		if cmd != nil {
			filtered = append(filtered, cmd)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return tea.Batch(filtered...)
}
