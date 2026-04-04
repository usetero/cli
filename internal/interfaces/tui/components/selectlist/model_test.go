package selectlist

import (
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

type testRow struct {
	title    string
	subtitle string
}

func (r testRow) FilterValue() string { return r.title + " " + r.subtitle }
func (r testRow) Title() string       { return r.title }
func (r testRow) Subtitle() string    { return r.subtitle }

var _ Row = testRow{}
var _ list.Item = testRow{}

func TestPreferredHeightIgnoresAssignedViewportHeight(t *testing.T) {
	m := New(theme.Default(), false, "")
	m.SetItems([]list.Item{
		testRow{title: "Quit", subtitle: "Close Tero"},
	})

	if got := m.PreferredHeight(80); got != 1 {
		t.Fatalf("expected intrinsic height 1 before sizing, got %d", got)
	}

	m.SetSize(80, 20)

	if got := m.PreferredHeight(80); got != 1 {
		t.Fatalf("expected intrinsic height 1 after assigned height, got %d", got)
	}
}

func TestPreferredHeightStaysStableWhenSelectionMoves(t *testing.T) {
	m := New(theme.Default(), false, "")
	m.SetItems([]list.Item{
		testRow{title: "One"},
		testRow{title: "Two"},
		testRow{title: "Three"},
	})

	want := m.PreferredHeight(80)
	if want != 3 {
		t.Fatalf("expected intrinsic height 3, got %d", want)
	}

	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})

	if got := m.PreferredHeight(80); got != want {
		t.Fatalf("expected intrinsic height to remain %d after selection movement, got %d", want, got)
	}
}
