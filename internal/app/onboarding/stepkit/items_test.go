package stepkit

import (
	"testing"

	"github.com/usetero/cli/internal/tea/components/remotelist"
)

func TestCastItems(t *testing.T) {
	t.Parallel()

	items := []remotelist.Item{castTestItem("a"), castTestItem("b")}
	got := CastItems[castTestItem](items)
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	if got[0] != "a" || got[1] != "b" {
		t.Fatalf("unexpected cast result: %#v", got)
	}
}

type castTestItem string

func (i castTestItem) FilterValue() string { return string(i) }
