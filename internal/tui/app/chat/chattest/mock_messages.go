// Package chattest provides test utilities for the chat package.
package chattest

import (
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/tui/app/chat"
)

// MockMessages is a test double for chat.Messages.
type MockMessages struct {
	InitFunc            func() tea.Cmd
	UpdateFunc          func(msg tea.Msg) tea.Cmd
	SetConversationFunc func(conversationID string) tea.Cmd
	RefreshFunc         func() tea.Cmd
	SetWidthFunc        func(width int)
	ItemsFunc           func() []chat.Item
	HasErrorFunc        func() bool
	ErrorFunc           func() error
	IsBusyFunc          func() bool
	CloseFunc           func() error
}

// Compile-time check that MockMessages implements chat.Messages.
var _ chat.Messages = (*MockMessages)(nil)

func (m *MockMessages) Init() tea.Cmd {
	if m.InitFunc != nil {
		return m.InitFunc()
	}
	return nil
}

func (m *MockMessages) Update(msg tea.Msg) tea.Cmd {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(msg)
	}
	return nil
}

func (m *MockMessages) SetConversation(conversationID string) tea.Cmd {
	if m.SetConversationFunc != nil {
		return m.SetConversationFunc(conversationID)
	}
	return nil
}

func (m *MockMessages) Refresh() tea.Cmd {
	if m.RefreshFunc != nil {
		return m.RefreshFunc()
	}
	return nil
}

func (m *MockMessages) SetWidth(width int) {
	if m.SetWidthFunc != nil {
		m.SetWidthFunc(width)
	}
}

func (m *MockMessages) Items() []chat.Item {
	if m.ItemsFunc != nil {
		return m.ItemsFunc()
	}
	return nil
}

func (m *MockMessages) HasError() bool {
	if m.HasErrorFunc != nil {
		return m.HasErrorFunc()
	}
	return false
}

func (m *MockMessages) Error() error {
	if m.ErrorFunc != nil {
		return m.ErrorFunc()
	}
	return nil
}

func (m *MockMessages) IsBusy() bool {
	if m.IsBusyFunc != nil {
		return m.IsBusyFunc()
	}
	return false
}

func (m *MockMessages) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
