package listdetail

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/tea/keymap"
)

// Controller encapsulates shared list/detail drawer navigation behavior.
type Controller struct {
	HasList func() bool

	IsDetail    func() bool
	CloseDetail func()

	GetListCursor func() int
	SetListCursor func(int)
	ListLen       func() int
	OnListSelect  func(index int) tea.Cmd

	GetDetailCursor func() int
	SetDetailCursor func(int)
	DetailLen       func() int
	OnDetailSelect  func() tea.Cmd
}

// Detail is the minimal contract for detail panes used by drawer navigation.
type Detail interface {
	Len() int
	Cursor() int
	SetCursor(int)
	Prompt() tea.Cmd
}

// New builds a standard list/detail controller from simple state hooks.
func New(
	hasList func() bool,
	getListCursor func() int,
	setListCursor func(int),
	listLen func() int,
	onListSelect func(index int) tea.Cmd,
	getDetail func() Detail,
	clearDetail func(),
) Controller {
	return Controller{
		HasList: hasList,
		IsDetail: func() bool {
			return getDetail() != nil
		},
		CloseDetail: clearDetail,
		GetListCursor: func() int {
			return getListCursor()
		},
		SetListCursor: setListCursor,
		ListLen:       listLen,
		OnListSelect:  onListSelect,
		GetDetailCursor: func() int {
			d := getDetail()
			if d == nil {
				return 0
			}
			return d.Cursor()
		},
		SetDetailCursor: func(v int) {
			d := getDetail()
			if d != nil {
				d.SetCursor(v)
			}
		},
		DetailLen: func() int {
			d := getDetail()
			if d == nil {
				return 0
			}
			return d.Len()
		},
		OnDetailSelect: func() tea.Cmd {
			d := getDetail()
			if d == nil {
				return nil
			}
			return d.Prompt()
		},
	}
}

// HandleKeyPress processes drawer navigation keys for list/detail views.
func (c Controller) HandleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	if c.HasList == nil || !c.HasList() {
		return nil
	}

	if c.IsDetail != nil && c.IsDetail() {
		return c.handleDetailKeyPress(msg)
	}
	return c.handleListKeyPress(msg)
}

func (c Controller) handleDetailKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case key.Matches(msg, keymap.DrawerBack):
		if c.CloseDetail != nil {
			c.CloseDetail()
		}
	case key.Matches(msg, keymap.DrawerUp):
		if c.GetDetailCursor == nil || c.SetDetailCursor == nil {
			return nil
		}
		cursor := c.GetDetailCursor()
		if cursor > 0 {
			c.SetDetailCursor(cursor - 1)
		}
	case key.Matches(msg, keymap.DrawerDown):
		if c.GetDetailCursor == nil || c.SetDetailCursor == nil || c.DetailLen == nil {
			return nil
		}
		cursor := c.GetDetailCursor()
		if cursor < c.DetailLen()-1 {
			c.SetDetailCursor(cursor + 1)
		}
	case key.Matches(msg, keymap.DrawerSelect):
		if c.OnDetailSelect != nil {
			return c.OnDetailSelect()
		}
	}
	return nil
}

func (c Controller) handleListKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case key.Matches(msg, keymap.DrawerUp):
		if c.GetListCursor == nil || c.SetListCursor == nil {
			return nil
		}
		cursor := c.GetListCursor()
		if cursor > 0 {
			c.SetListCursor(cursor - 1)
		}
	case key.Matches(msg, keymap.DrawerDown):
		if c.GetListCursor == nil || c.SetListCursor == nil || c.ListLen == nil {
			return nil
		}
		cursor := c.GetListCursor()
		if cursor < c.ListLen()-1 {
			c.SetListCursor(cursor + 1)
		}
	case key.Matches(msg, keymap.DrawerSelect):
		if c.OnListSelect != nil && c.GetListCursor != nil {
			return c.OnListSelect(c.GetListCursor())
		}
	}
	return nil
}

// ClampCursor keeps cursor within [0, length-1] when length > 0.
func ClampCursor(cursor, length int) int {
	if length <= 0 {
		return 0
	}
	if cursor < 0 {
		return 0
	}
	if cursor >= length {
		return length - 1
	}
	return cursor
}
