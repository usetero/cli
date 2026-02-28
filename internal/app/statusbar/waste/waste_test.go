package waste

import (
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
