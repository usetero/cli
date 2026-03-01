package tabpoll

import (
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestUpdatePollCycle_PollStartsFetch(t *testing.T) {
	t.Parallel()

	fetching := false
	fetchCmd := func() tea.Msg { return "fetch" }
	pollCmd := func() tea.Msg { return "poll" }
	var applied bool

	cmd, handled := UpdatePollCycle[string](
		PollMsg{Source: "waste"},
		"waste",
		true,
		&fetching,
		fetchCmd,
		pollCmd,
		func(string) { applied = true },
	)

	if !handled {
		t.Fatalf("expected handled=true")
	}
	if !fetching {
		t.Fatalf("expected fetching=true")
	}
	if cmd == nil {
		t.Fatalf("expected non-nil cmd")
	}
	if applied {
		t.Fatalf("did not expect applyData on PollMsg")
	}

	msg := cmd()
	if _, ok := msg.(tea.BatchMsg); !ok {
		t.Fatalf("expected tea.BatchMsg, got %T", msg)
	}
}

func TestUpdatePollCycle_PollWhileFetchingOnlySchedulesPoll(t *testing.T) {
	t.Parallel()

	fetching := true
	fetchCmd := func() tea.Msg { return "fetch" }
	pollCmd := func() tea.Msg { return "poll" }

	cmd, handled := UpdatePollCycle[string](
		PollMsg{Source: "waste"},
		"waste",
		true,
		&fetching,
		fetchCmd,
		pollCmd,
		func(string) {},
	)

	if !handled {
		t.Fatalf("expected handled=true")
	}
	if !fetching {
		t.Fatalf("expected fetching to remain true")
	}
	if cmd == nil {
		t.Fatalf("expected non-nil cmd")
	}

	if got := cmd(); got != "poll" {
		t.Fatalf("expected poll msg, got %v", got)
	}
}

func TestUpdatePollCycle_DataAppliesAndClearsFetching(t *testing.T) {
	t.Parallel()

	fetching := true
	var gotData string

	cmd, handled := UpdatePollCycle[string](
		DataMsg[string]{Data: "ok"},
		"waste",
		true,
		&fetching,
		nil,
		nil,
		func(data string) { gotData = data },
	)

	if !handled {
		t.Fatalf("expected handled=true")
	}
	if cmd != nil {
		t.Fatalf("expected nil cmd")
	}
	if fetching {
		t.Fatalf("expected fetching=false")
	}
	if gotData != "ok" {
		t.Fatalf("expected applyData with ok, got %q", gotData)
	}
}

func TestUpdatePollCycle_DataErrorSkipsApply(t *testing.T) {
	t.Parallel()

	fetching := true
	called := false

	cmd, handled := UpdatePollCycle[string](
		DataMsg[string]{Err: errors.New("boom")},
		"waste",
		true,
		&fetching,
		nil,
		nil,
		func(string) { called = true },
	)

	if !handled {
		t.Fatalf("expected handled=true")
	}
	if cmd != nil {
		t.Fatalf("expected nil cmd")
	}
	if fetching {
		t.Fatalf("expected fetching=false")
	}
	if called {
		t.Fatalf("did not expect applyData to be called")
	}
}

func TestUpdatePollCycle_UnhandledMessage(t *testing.T) {
	t.Parallel()

	fetching := false
	cmd, handled := UpdatePollCycle[string](
		tea.KeyPressMsg{},
		"waste",
		true,
		&fetching,
		nil,
		nil,
		func(string) {},
	)

	if handled {
		t.Fatalf("expected handled=false")
	}
	if cmd != nil {
		t.Fatalf("expected nil cmd")
	}
}

func TestTickEmitsSource(t *testing.T) {
	t.Parallel()

	cmd := Tick("quality", time.Millisecond)
	msg := cmd()
	poll, ok := msg.(PollMsg)
	if !ok {
		t.Fatalf("expected PollMsg, got %T", msg)
	}
	if poll.Source != "quality" {
		t.Fatalf("expected source quality, got %q", poll.Source)
	}
}
