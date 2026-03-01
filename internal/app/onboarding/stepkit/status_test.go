package stepkit

import "testing"

func TestAlwaysVisible(t *testing.T) {
	t.Parallel()

	if AlwaysVisible() {
		t.Fatal("expected AlwaysVisible to return hidden=false")
	}
}

func TestStaticStatus(t *testing.T) {
	t.Parallel()

	got := StaticStatus("Title", "Details")
	if got.Title != "Title" || got.Details != "Details" {
		t.Fatalf("unexpected status: %#v", got)
	}
}
