package policytab

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/app/statusbar/listdetail"
	"github.com/usetero/cli/internal/app/statusbar/tabpoll"
	"github.com/usetero/cli/internal/sqlite"
)

const defaultPollInterval = 2 * time.Second

// Base holds shared state/lifecycle behavior for policy-like status bar tabs.
type Base struct {
	db        sqlite.DB
	source    string
	hasData   bool
	lastState string
	fetching  bool
	cursor    int
}

func New(source string) Base {
	return Base{source: source}
}

func (b *Base) SetDB(db sqlite.DB) tea.Cmd {
	b.db = db
	return b.poll()
}

func (b *Base) Init() tea.Cmd {
	if b.db == nil {
		return nil
	}
	return b.poll()
}

func (b *Base) DB() sqlite.DB           { return b.db }
func (b *Base) HasData() bool           { return b.hasData }
func (b *Base) Cursor() int             { return b.cursor }
func (b *Base) SetCursor(v int)         { b.cursor = v }
func (b *Base) SetHasData(v bool)       { b.hasData = v }
func (b *Base) HasList(length int) bool { return b.hasData && length > 0 }

func (b *Base) poll() tea.Cmd {
	return tabpoll.Tick(b.source, defaultPollInterval)
}

// UpdatePoll handles the shared PollMsg/DataMsg cycle.
func UpdatePoll[T any](b *Base, msg tea.Msg, fetchData tea.Cmd, applyData func(data T)) (tea.Cmd, bool) {
	return tabpoll.UpdatePollCycle(
		msg,
		b.source,
		b.db != nil,
		&b.fetching,
		fetchData,
		b.poll(),
		applyData,
	)
}

// ApplyIfChanged applies state updates on key changes and clamps cursor.
func (b *Base) ApplyIfChanged(nextState string, listLen int, apply func()) bool {
	return tabpoll.ApplyIfChanged(&b.lastState, nextState, &b.cursor, listLen, apply)
}

// NavController returns standard list/detail drawer navigation wiring.
func (b *Base) NavController(
	listLen func() int,
	onListSelect func(index int) tea.Cmd,
	getDetail func() listdetail.Detail,
	clearDetail func(),
) listdetail.Controller {
	return listdetail.New(
		func() bool { return b.HasList(listLen()) },
		func() int { return b.cursor },
		func(v int) { b.cursor = v },
		listLen,
		onListSelect,
		getDetail,
		clearDetail,
	)
}
