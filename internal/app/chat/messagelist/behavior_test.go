package messagelist

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/app/chat/messagelist/round"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools/action"
	"github.com/usetero/cli/internal/app/chat/msgs"
	chat "github.com/usetero/cli/internal/chat"
	chattools "github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/domain"
	domaintools "github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/tea/teatest"
)

func addCompletedRound(t *testing.T, m *Model, turnID domain.MessageID, text string) {
	t.Helper()

	m.StartTurn("conv-1", "acct-1", turnID, msgs.UserSubmittedInput{Text: "prompt " + string(turnID)}, nil, nil)
	m.Update(msgs.StreamCompleted{
		TurnID:     turnID,
		StopReason: "end_turn",
		Message: domain.Message{
			ID:         "asst-" + turnID,
			StopReason: "end_turn",
			Content: []domain.Block{
				{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: text}},
			},
		},
	})
}

func seedHistoryWithActiveRound(t *testing.T, height int) *Model {
	t.Helper()

	m := newStreamingMessageList(t)
	m.SetSize(80, height)

	for i := range 6 {
		id := domain.MessageID(fmt.Sprintf("user-%d", i+1))
		addCompletedRound(t, m, id, fmt.Sprintf("history %d", i+1))
	}

	m.StartTurn("conv-1", "acct-1", "user-live", msgs.UserSubmittedInput{Text: "live"}, nil, nil)
	return m
}

type toggleSpyBlock struct {
	text      string
	toggleCnt int
}

func (b *toggleSpyBlock) View() string           { return b.text }
func (b *toggleSpyBlock) Height() int            { return 1 }
func (b *toggleSpyBlock) Update(tea.Msg) tea.Cmd { return nil }
func (b *toggleSpyBlock) SetWidth(int)           {}
func (b *toggleSpyBlock) SetFocused(bool)        {}
func (b *toggleSpyBlock) Focused() bool          { return false }
func (b *toggleSpyBlock) Kind() block.Kind       { return block.KindAssistantText }
func (b *toggleSpyBlock) Toggle(int)             { b.toggleCnt++ }

func TestBehavior_CancelledRoundIgnoresStaleAssistantUpdates(t *testing.T) {
	t.Parallel()

	m := newStreamingMessageList(t)

	m.StartTurn("conv-1", "acct-1", "user-1", msgs.UserSubmittedInput{Text: "first"}, nil, nil)
	m.CancelActiveRound()
	m.StartTurn("conv-1", "acct-1", "user-2", msgs.UserSubmittedInput{Text: "second"}, nil, nil)

	m.Update(msgs.AssistantContentUpdated{
		TurnID: "user-1",
		Message: domain.Message{
			ID: "asst-stale",
			Content: []domain.Block{
				{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "stale text"}},
			},
		},
	})

	if len(m.rounds) != 2 {
		t.Fatalf("round count=%d, want 2", len(m.rounds))
	}
	if m.rounds[0].State() != round.StateCancelled {
		t.Fatalf("round[0] state=%v, want cancelled", m.rounds[0].State())
	}
	if !m.rounds[1].IsActive() {
		t.Fatalf("round[1] should still be active")
	}
	if strings.Contains(m.View(), "stale text") {
		t.Fatalf("stale update should not be rendered in view")
	}
}

func TestBehavior_StreamUpdateScrollPolicy(t *testing.T) {
	t.Parallel()

	t.Run("at bottom sticks to bottom on assistant updates", func(t *testing.T) {
		t.Parallel()
		m := seedHistoryWithActiveRound(t, 8)
		m.vp.ScrollToBottom()

		m.Update(msgs.AssistantContentUpdated{
			TurnID: "user-live",
			Message: domain.Message{
				ID: "asst-live",
				Content: []domain.Block{
					{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "live update"}},
				},
			},
		})

		if !m.vp.AtBottom() {
			t.Fatalf("expected to remain at bottom after update")
		}
	})

	t.Run("scrolled up does not get yanked to bottom", func(t *testing.T) {
		t.Parallel()
		m := seedHistoryWithActiveRound(t, 8)
		m.vp.ScrollToBottom()
		m.vp.ScrollBy(-4)
		m.vp.UpdateFocusFromScroll()

		if m.vp.AtBottom() {
			t.Fatalf("precondition failed: expected scrolled-up viewport")
		}
		beforeIdx, beforeLine := m.vp.Offset()

		m.Update(msgs.AssistantContentUpdated{
			TurnID: "user-live",
			Message: domain.Message{
				ID: "asst-live",
				Content: []domain.Block{
					{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "live update"}},
				},
			},
		})

		if m.vp.AtBottom() {
			t.Fatalf("viewport should stay scrolled up after update")
		}
		afterIdx, afterLine := m.vp.Offset()
		if beforeIdx != afterIdx || beforeLine != afterLine {
			t.Fatalf("expected offset stability while scrolled up: before=(%d,%d) after=(%d,%d)", beforeIdx, beforeLine, afterIdx, afterLine)
		}
	})
}

func TestBehavior_MouseReleaseActionPolicy(t *testing.T) {
	t.Parallel()

	newToggleModel := func(t *testing.T) (*Model, *toggleSpyBlock, *toggleSpyBlock) {
		t.Helper()
		m := newStreamingMessageList(t)
		m.SetSize(80, 8)
		m.SetOrigin(0, 0)
		addCompletedRound(t, m, "seed-round", "seed")

		b0 := &toggleSpyBlock{text: "alpha"}
		b1 := &toggleSpyBlock{text: "beta"}
		m.blocks = []blockEntry{
			{block: b0, roundIndex: 0},
			{block: b1, roundIndex: 0},
		}
		// 2 blocks with 1-line heights and 1-line gap between them:
		// block 0 at y=0, gap at y=1, block 1 at y=2.
		m.vp.SetItems([]int{1, 1}, []int{0, 1})
		m.vp.SetTrailingHeight(0)
		m.vp.ScrollToTop()
		return m, b0, b1
	}

	t.Run("plain click triggers toggle", func(t *testing.T) {
		t.Parallel()
		m, b0, b1 := newToggleModel(t)

		m.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 2, Y: 0})
		m.Update(tea.MouseReleaseMsg{Button: tea.MouseLeft, X: 2, Y: 0})

		if b0.toggleCnt != 1 {
			t.Fatalf("expected first block toggle once, got %d", b0.toggleCnt)
		}
		if b1.toggleCnt != 0 {
			t.Fatalf("expected second block untouched, got %d", b1.toggleCnt)
		}
	})

	t.Run("drag selection across blocks does not toggle", func(t *testing.T) {
		t.Parallel()
		m, b0, b1 := newToggleModel(t)

		m.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 2, Y: 0})
		m.Update(tea.MouseMotionMsg{Button: tea.MouseLeft, X: 2, Y: 2})
		cmd := m.Update(tea.MouseReleaseMsg{Button: tea.MouseLeft, X: 2, Y: 2})

		if cmd == nil {
			t.Fatalf("expected copy command on drag highlight release")
		}
		if b0.toggleCnt != 0 || b1.toggleCnt != 0 {
			t.Fatalf("expected no toggles on drag copy, got b0=%d b1=%d", b0.toggleCnt, b1.toggleCnt)
		}
	})
}

func TestBehavior_StaleToolCompletedIgnored(t *testing.T) {
	t.Parallel()

	m := newStreamingMessageList(t)
	m.StartTurn("conv-1", "acct-1", "user-1", msgs.UserSubmittedInput{Text: "first"}, nil, nil)
	m.CancelActiveRound()
	m.StartTurn("conv-1", "acct-1", "user-2", msgs.UserSubmittedInput{Text: "second"}, nil, nil)

	if len(m.rounds) != 2 {
		t.Fatalf("round count=%d, want 2", len(m.rounds))
	}
	if m.rounds[0].State() != round.StateCancelled {
		t.Fatalf("round[0] state=%v, want cancelled", m.rounds[0].State())
	}
	if !m.rounds[1].IsActive() {
		t.Fatalf("round[1] should be active before stale tool completion")
	}

	cmd := m.Update(msgs.ToolCompleted{
		TurnID:    "user-1",
		ToolUseID: "tool-old",
		Result:    domaintools.Result{ToolUseID: "tool-old"},
	})

	if cmd != nil {
		t.Fatalf("expected stale ToolCompleted to be ignored, got non-nil cmd")
	}
	if !m.rounds[1].IsActive() {
		t.Fatalf("round[1] should remain active after stale tool completion")
	}
}

func TestBehavior_ToolResultsStayBoundToOriginalBlock(t *testing.T) {
	t.Parallel()

	type toolInput struct {
		Name string `json:"name"`
	}

	actionTool := chattools.NewActionTool(
		chat.Tool{Name: "set_service_enabled"},
		func(input json.RawMessage) (domaintools.Result, error) {
			var in toolInput
			if err := json.Unmarshal(input, &in); err != nil {
				return domaintools.Result{}, err
			}
			return domaintools.Result{
				Content: map[string]any{
					"name": in.Name,
				},
			}, nil
		},
		action.Config{
			DisplayName: func(_ json.RawMessage) string { return "Enable Service" },
			Status:      func(_ json.RawMessage) string { return "Enabling" },
			Result: func(result domaintools.Result) string {
				name, _ := result.Content["name"].(string)
				return name + " enabled"
			},
		},
	)
	registry := chattools.NewRegistry(nil, nil, map[string]chattools.ActionTool{
		"set_service_enabled": actionTool,
	})

	m := newStreamingMessageList(t)
	m.toolRegistry = registry

	m.StartTurn("conv-1", "acct-1", "user-1", msgs.UserSubmittedInput{Text: "enable"}, nil, nil)

	cmd1 := m.Update(msgs.AssistantContentUpdated{
		TurnID: "user-1",
		Message: domain.Message{
			ID: "asst-1",
			Content: []domain.Block{
				{
					Index: 0,
					Type:  domain.BlockTypeToolUse,
					ToolUse: &domain.ToolUse{
						ID:            "tool-1",
						Name:          "set_service_enabled",
						Input:         json.RawMessage(`{"name":"alpha"}`),
						InputComplete: true,
					},
				},
			},
		},
	})
	teatest.DrainCmds(m.Update, cmd1, 64)

	viewAfterFirst := m.View()
	if !strings.Contains(viewAfterFirst, "alpha enabled") {
		t.Fatalf("expected first result in view, got:\n%s", viewAfterFirst)
	}

	cmd2 := m.Update(msgs.AssistantContentUpdated{
		TurnID: "user-1",
		Message: domain.Message{
			ID: "asst-1",
			Content: []domain.Block{
				{
					Index: 0,
					Type:  domain.BlockTypeToolUse,
					ToolUse: &domain.ToolUse{
						ID:            "tool-1",
						Name:          "set_service_enabled",
						Input:         json.RawMessage(`{"name":"alpha"}`),
						InputComplete: true,
					},
				},
				{
					Index: 1,
					Type:  domain.BlockTypeToolUse,
					ToolUse: &domain.ToolUse{
						ID:            "tool-2",
						Name:          "set_service_enabled",
						Input:         json.RawMessage(`{"name":"beta"}`),
						InputComplete: true,
					},
				},
			},
		},
	})
	teatest.DrainCmds(m.Update, cmd2, 128)

	viewAfterSecond := m.View()
	if strings.Count(viewAfterSecond, "alpha enabled") != 1 {
		t.Fatalf("expected alpha result to remain exactly once, got:\n%s", viewAfterSecond)
	}
	if strings.Count(viewAfterSecond, "beta enabled") != 1 {
		t.Fatalf("expected beta result exactly once, got:\n%s", viewAfterSecond)
	}
}
