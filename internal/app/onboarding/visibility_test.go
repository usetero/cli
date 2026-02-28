package onboarding

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	appmsg "github.com/usetero/cli/internal/app/onboarding/msgs"
)

func TestViewDoesNotHideByGateAlone(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m.SetSize(80, 20)
	m.gate = GateRuntimeInit
	m.step = fixedTestStep{view: "runtime detail should be visible"}

	view := m.View()
	if !strings.Contains(view, "runtime detail should be visible") {
		t.Fatalf("expected gate alone not to hide step view, got: %q", view)
	}
}

func TestViewUsesHiddenStepStatus(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m.SetSize(80, 20)
	m.gate = GateRoleSelect // visible by default, hidden should come from provider
	m.step = hiddenTestStep{
		fixedTestStep: fixedTestStep{view: "step view should be hidden"},
		hidden:        true,
		status: appmsg.StepStatus{
			Title:   "Loading",
			Details: "Loading dynamic data...",
		},
	}

	view := m.View()
	if !strings.Contains(view, "Loading dynamic data...") {
		t.Fatalf("expected provider status text in hidden view, got: %q", view)
	}
	if strings.Contains(view, "step view should be hidden") {
		t.Fatalf("expected hidden provider to suppress step view")
	}
}

func TestViewUsesOverriddenEmbeddedStatus(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m.SetSize(80, 20)
	m.gate = GateDatadogCheck
	m.step = hiddenStatusTestStep{
		hiddenTestStep: hiddenTestStep{
			fixedTestStep: fixedTestStep{view: "step view should be hidden"},
			hidden:        true,
			status: appmsg.StepStatus{
				Title:   "Fallback",
				Details: "fallback status text",
			},
		},
		status: appmsg.StepStatus{
			Title:   "Datadog setup",
			Details: "Checking Datadog configuration...",
		},
	}

	view := m.View()
	if !strings.Contains(view, "Datadog setup") {
		t.Fatalf("expected structured status title in hidden view, got: %q", view)
	}
	if !strings.Contains(view, "Checking Datadog configuration...") {
		t.Fatalf("expected structured status details in hidden view, got: %q", view)
	}
	if strings.Contains(view, "fallback status text") {
		t.Fatalf("expected structured status to override fallback status text")
	}
}

func TestViewHiddenStatusNeverFallsBackToGenericTitle(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m.SetSize(80, 20)
	m.gate = GateRuntimeInit
	m.step = hiddenTestStep{
		fixedTestStep: fixedTestStep{view: "step view should be hidden"},
		hidden:        true,
		status: appmsg.StepStatus{
			Title:   "Runtime setup",
			Details: "Initializing account runtime...",
		},
	}

	view := m.View()
	if strings.Contains(view, "Getting ready") {
		t.Fatalf("expected hidden view to use provider title, got generic title: %q", view)
	}
	if !strings.Contains(view, "Runtime setup") {
		t.Fatalf("expected provider title in hidden view, got: %q", view)
	}
}

type fixedTestStep struct {
	view string
}

func (s fixedTestStep) Init() tea.Cmd              { return nil }
func (s fixedTestStep) Update(msg tea.Msg) tea.Cmd { return nil }
func (s fixedTestStep) View() string               { return s.view }
func (s fixedTestStep) SetSize(width, height int)  {}
func (s fixedTestStep) ShortHelp() []key.Binding   { return nil }
func (s fixedTestStep) Hidden() bool               { return false }
func (s fixedTestStep) Status() appmsg.StepStatus  { return appmsg.StepStatus{} }

type hiddenTestStep struct {
	fixedTestStep
	hidden bool
	status appmsg.StepStatus
}

func (s hiddenTestStep) Hidden() bool { return s.hidden }
func (s hiddenTestStep) Status() appmsg.StepStatus {
	return s.status
}

type hiddenStatusTestStep struct {
	hiddenTestStep
	status appmsg.StepStatus
}

func (s hiddenStatusTestStep) Status() appmsg.StepStatus { return s.status }
