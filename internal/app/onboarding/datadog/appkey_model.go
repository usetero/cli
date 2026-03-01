package datadog

import (
	"context"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	api "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/input"
)

// AppKeyModel handles Datadog Application key entry and account creation.
type AppKeyModel struct {
	ctx      context.Context
	theme    styles.Theme
	services api.APIServices
	scope    log.Scope
	account  domain.Account
	site     domain.DatadogSite
	apiKey   string

	input    *input.Model
	creating bool
	err      error
	width    int
	height   int
}

// NewAppKey creates a new App key entry step.
func NewAppKey(
	ctx context.Context,
	theme styles.Theme,
	account domain.Account,
	site domain.DatadogSite,
	apiKey string,
	services api.APIServices,
	scope log.Scope,
) *AppKeyModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if site == "" {
		panic("site is empty")
	}
	if apiKey == "" {
		panic("apiKey is empty")
	}

	inp := input.New(theme)
	inp.SetPlaceholder("Datadog Application Key")
	inp.SetCharLimit(64)
	inp.SetEchoMode(textinput.EchoPassword)
	inp.SetEchoCharacter('•')

	return &AppKeyModel{
		ctx:      ctx,
		theme:    theme,
		services: services,
		scope:    scope,
		account:  account,
		site:     site,
		apiKey:   apiKey,
		input:    inp,
	}
}

// Init focuses the input.
func (m *AppKeyModel) Init() tea.Cmd {
	return m.input.Focus()
}

// SetSize updates dimensions.
func (m *AppKeyModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.input.SetWidth(width)
}

// ShortHelp returns the key bindings for the short help view.
func (m *AppKeyModel) ShortHelp() []key.Binding {
	if m.creating {
		return nil
	}
	bindings := m.input.ShortHelp()
	bindings = append(bindings,
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "connect")),
	)
	return bindings
}

func appKeyErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return "Failed to connect Datadog. Please try again."
	}

	// Strip common GraphQL wrapper prefixes so users see the actionable backend message.
	msg = strings.TrimSpace(strings.TrimPrefix(msg, "graphql:"))

	lower := strings.ToLower(msg)
	if strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "deadline exceeded") ||
		strings.Contains(lower, "connection refused") ||
		strings.Contains(lower, "network is unreachable") {
		return "Could not reach Datadog. Check your connection and try again."
	}

	return msg
}
