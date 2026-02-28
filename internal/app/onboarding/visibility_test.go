package onboarding

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func TestDisplayPolicyForGate(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)

	runtimePolicy := m.displayPolicyForGate(GateRuntimeInit)
	if !runtimePolicy.hidden {
		t.Fatalf("runtime init gate should be hidden by default")
	}
	if runtimePolicy.status == "" {
		t.Fatalf("runtime init gate should provide default status text")
	}

	rolePolicy := m.displayPolicyForGate(GateRoleSelect)
	if rolePolicy.hidden {
		t.Fatalf("role select should be visible by default")
	}
}

func TestViewUsesGateDisplayPolicyHiddenState(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m.SetSize(80, 20)
	m.gate = GateRuntimeInit
	m.step = fixedTestStep{view: "runtime detail that should be hidden"}

	view := m.View()
	if !strings.Contains(view, "Getting ready") {
		t.Fatalf("expected hidden gate heading in view, got: %q", view)
	}
	if strings.Contains(view, "runtime detail that should be hidden") {
		t.Fatalf("expected hidden gate to suppress step detail")
	}
}

func TestViewUsesVisibilityProviderOverride(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m.SetSize(80, 20)
	m.gate = GateRoleSelect // visible by default, hidden should come from provider
	m.step = hiddenTestStep{
		fixedTestStep: fixedTestStep{view: "step view should be hidden"},
		hidden:        true,
		status:        "Loading dynamic data...",
	}

	view := m.View()
	if !strings.Contains(view, "Loading dynamic data...") {
		t.Fatalf("expected provider status text in hidden view, got: %q", view)
	}
	if strings.Contains(view, "step view should be hidden") {
		t.Fatalf("expected hidden provider to suppress step view")
	}
}

type fixedTestStep struct {
	view string
}

func (s fixedTestStep) Init() tea.Cmd                { return nil }
func (s fixedTestStep) Update(msg tea.Msg) tea.Cmd   { return nil }
func (s fixedTestStep) View() string                 { return s.view }
func (s fixedTestStep) SetSize(width, height int)    {}
func (s fixedTestStep) ShortHelp() []key.Binding     { return nil }

type hiddenTestStep struct {
	fixedTestStep
	hidden bool
	status string
}

func (s hiddenTestStep) Hidden() bool       { return s.hidden }
func (s hiddenTestStep) StatusText() string { return s.status }
