package screen

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type routerTestMsg struct {
	value string
}

type childID string

const (
	childIDA childID = "a"
	childIDB childID = "b"
	childIDC childID = "child"
)

type routerChildModel struct {
	name       string
	updates    int
	lastMsg    tea.Msg
	width      int
	height     int
	help       []key.Binding
	returnCmd  tea.Cmd
	returnNext *routerChildModel
}

func (m *routerChildModel) Init() tea.Cmd { return nil }

func (m *routerChildModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.updates++
	m.lastMsg = msg
	if m.returnNext != nil {
		return m.returnNext, m.returnCmd
	}
	return m, m.returnCmd
}

func (m *routerChildModel) View() tea.View { return tea.NewView(m.name) }

func (m *routerChildModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *routerChildModel) ShortHelp() []key.Binding { return m.help }

func TestRouter_ActivateOnlyForwardsToSingleChild(t *testing.T) {
	a := &routerChildModel{name: "a"}
	b := &routerChildModel{name: "b"}
	r := &Router[childID]{}
	r.Register(childIDA, a)
	r.Register(childIDB, b)
	if ok := r.ActivateOnly(childIDB); !ok {
		t.Fatal("expected child b to be registered")
	}

	cmd := r.Forward(routerTestMsg{value: "x"})
	if cmd != nil {
		t.Fatal("did not expect command")
	}
	if a.updates != 0 {
		t.Fatalf("expected child a not to update, got %d", a.updates)
	}
	if b.updates != 1 {
		t.Fatalf("expected child b to update once, got %d", b.updates)
	}
}

func TestRouter_MultiActiveForwardsToAll(t *testing.T) {
	a := &routerChildModel{name: "a"}
	b := &routerChildModel{name: "b"}
	r := &Router[childID]{}
	r.Register(childIDA, a)
	r.Register(childIDB, b)
	r.SetActive(childIDA, true)
	r.SetActive(childIDB, true)

	r.Forward(routerTestMsg{value: "x"})

	if a.updates != 1 || b.updates != 1 {
		t.Fatalf("expected both children to update once, got a=%d b=%d", a.updates, b.updates)
	}
}

func TestRouter_ForwardStoresReturnedModelReplacement(t *testing.T) {
	next := &routerChildModel{name: "next"}
	orig := &routerChildModel{name: "orig", returnNext: next}
	r := &Router[childID]{}
	r.Register(childIDC, orig)
	r.SetActive(childIDC, true)

	r.Forward(routerTestMsg{value: "x"})

	got, ok := r.Model(childIDC).(*routerChildModel)
	if !ok {
		t.Fatalf("expected *routerChildModel, got %T", r.Model(childIDC))
	}
	if got != next {
		t.Fatal("expected router to store replacement model")
	}
}

func TestRouter_LiftTransformsChildCmd(t *testing.T) {
	a := &routerChildModel{
		name: "a",
		returnCmd: func() tea.Msg {
			return routerTestMsg{value: "child"}
		},
	}
	r := &Router[childID]{}
	r.Register(childIDA, a)
	r.SetLift(childIDA, func(cmd tea.Cmd) tea.Cmd {
		if cmd == nil {
			return nil
		}
		return func() tea.Msg {
			msg := cmd()
			typed, ok := msg.(routerTestMsg)
			if !ok {
				return msg
			}
			return routerTestMsg{value: "lifted:" + typed.value}
		}
	})
	r.SetActive(childIDA, true)

	cmd := r.Forward(routerTestMsg{value: "x"})
	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	typed, ok := msg.(routerTestMsg)
	if !ok {
		t.Fatalf("expected routerTestMsg, got %T", msg)
	}
	if typed.value != "lifted:child" {
		t.Fatalf("unexpected lifted message value %q", typed.value)
	}
}

func TestRouter_SetSizeAndShortHelpFromActiveChildren(t *testing.T) {
	helpA := key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "alpha"))
	helpB := key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "beta"))
	a := &routerChildModel{name: "a", help: []key.Binding{helpA}}
	b := &routerChildModel{name: "b", help: []key.Binding{helpB}}
	r := &Router[childID]{}
	r.Register(childIDA, a)
	r.Register(childIDB, b)
	r.SetActive(childIDA, true)
	r.SetActive(childIDB, false)

	r.SetSize(80, 24)
	if a.width != 80 || a.height != 24 {
		t.Fatalf("expected active child size to update, got %dx%d", a.width, a.height)
	}
	if b.width != 0 || b.height != 0 {
		t.Fatalf("expected inactive child size unchanged, got %dx%d", b.width, b.height)
	}

	help := r.ShortHelp()
	if len(help) != 1 {
		t.Fatalf("expected one help binding, got %d", len(help))
	}
	if help[0].Help().Desc != "alpha" {
		t.Fatalf("unexpected help binding %+v", help[0].Help())
	}

	r.SetSizeAll(100, 40)
	if b.width != 100 || b.height != 40 {
		t.Fatalf("expected SetSizeAll to size inactive child, got %dx%d", b.width, b.height)
	}
}
