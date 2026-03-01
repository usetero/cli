package stepkit

import (
	"testing"

	"charm.land/bubbles/v2/key"

	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/remotelist"
)

func TestRemoteListShortHelpLoading(t *testing.T) {
	t.Parallel()

	list := remotelist.New(styles.NewTheme(false), "Loading...")
	got := RemoteListShortHelp(list)
	if got != nil {
		t.Fatalf("expected nil help while loading, got %v", got)
	}
}

func TestRemoteListShortHelpError(t *testing.T) {
	t.Parallel()

	list := remotelist.New(styles.NewTheme(false), "Loading...")
	_ = list.Update(remotelist.LoadResult{Err: assertError{}})

	got := RemoteListShortHelp(list)
	if len(got) != 1 || got[0].Help().Key != "r" {
		t.Fatalf("expected retry binding, got %#v", got)
	}
}

func TestRemoteListShortHelpReadyIncludesExtras(t *testing.T) {
	t.Parallel()

	list := remotelist.New(styles.NewTheme(false), "Loading...")
	_ = list.Update(remotelist.LoadResult{})

	extra := key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm"))
	got := RemoteListShortHelp(list, extra)
	if len(got) == 0 {
		t.Fatalf("expected non-empty help")
	}
	last := got[len(got)-1]
	if last.Help().Key != "enter" {
		t.Fatalf("expected extra binding to be appended, got %q", last.Help().Key)
	}
}

type assertError struct{}

func (assertError) Error() string { return "boom" }
