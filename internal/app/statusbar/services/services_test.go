package services

import (
	"errors"
	"testing"

	"github.com/usetero/cli/internal/app/statusbar/tabpoll"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
	"github.com/usetero/cli/internal/styles"
)

func TestUpdatePollSourceFiltering(t *testing.T) {
	m := New(styles.NewTheme(true), logtest.NewScope(t))
	db := sqlitetest.OpenBareDB(t)
	m.SetDB(db)

	if cmd := m.Update(tabpoll.PollMsg{Source: "other"}); cmd != nil {
		t.Fatalf("expected foreign poll source to be ignored")
	}

	if cmd := m.Update(tabpoll.PollMsg{Source: pollSource}); cmd == nil {
		t.Fatalf("expected own poll source to schedule fetch")
	}
}

func TestUpdateDoesNotStartOverlappingFetch(t *testing.T) {
	m := New(styles.NewTheme(true), logtest.NewScope(t))
	db := sqlitetest.OpenBareDB(t)
	m.SetDB(db)

	if cmd := m.Update(tabpoll.PollMsg{Source: pollSource}); cmd == nil {
		t.Fatalf("expected poll to schedule fetch")
	}
	if !m.fetching {
		t.Fatalf("expected fetch to be marked in-flight")
	}

	if cmd := m.Update(tabpoll.PollMsg{Source: pollSource}); cmd == nil {
		t.Fatalf("expected second poll to keep poll loop alive")
	}
	if !m.fetching {
		t.Fatalf("expected in-flight flag to remain set while fetch is outstanding")
	}

	m.Update(tabpoll.DataMsg[fetchedData]{Err: errors.New("boom")})
	if m.fetching {
		t.Fatalf("expected in-flight flag to clear when data response arrives")
	}
}
