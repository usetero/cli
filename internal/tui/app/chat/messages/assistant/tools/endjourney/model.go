package endjourney

import (
	"encoding/json"

	"github.com/usetero/cli/internal/log"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/chat/messages/assistant/tools"
	appendjourney "github.com/usetero/cli/internal/tui/app/tools/endjourney"
)

// Model renders and executes an end_journey tool.
type Model struct {
	theme    *styles.Theme
	logger   log.Logger
	use      *domain.ToolUse
	executor *appendjourney.Tool
	result   *domain.ToolResult
}

// Compile-time interface check
var _ tools.Body = (*Model)(nil)

// New creates a new end_journey model.
func New(theme *styles.Theme, logger log.Logger, use *domain.ToolUse, executor *appendjourney.Tool) *Model {
	return &Model{
		theme:    theme,
		logger:   logger,
		use:      use,
		executor: executor,
	}
}

// Init executes the tool. Returns a Cmd that produces ResultMsg.
func (m *Model) Init() tea.Cmd {
	use := m.use
	executor := m.executor
	logger := m.logger

	return func() tea.Msg {
		logger.Info("ending journey")

		result, err := executor.Execute(use.Input)

		if err != nil {
			logger.Error("journey end failed", "error", err)
			return tools.ResultMsg{
				ToolUseID: use.ID,
				Result: &domain.ToolResult{
					ToolUseID: use.ID,
					IsError:   true,
					Error:     err.Error(),
				},
			}
		}

		logger.Info("journey ended")

		content, err := json.Marshal(result)
		if err != nil {
			return tools.ResultMsg{
				ToolUseID: use.ID,
				Result: &domain.ToolResult{
					ToolUseID: use.ID,
					IsError:   true,
					Error:     "failed to marshal result: " + err.Error(),
				},
			}
		}

		return tools.ResultMsg{
			ToolUseID: use.ID,
			Result: &domain.ToolResult{
				ToolUseID: use.ID,
				Content:   content,
			},
		}
	}
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tools.ResultMsg:
		if msg.ToolUseID == m.use.ID {
			m.result = msg.Result
		}
	}
	return nil
}

// Result returns the tool result.
func (m *Model) Result() *domain.ToolResult {
	return m.result
}

// Render returns the rendered body.
func (m *Model) Render(width int) string {
	colors := m.theme.Colors

	if m.result == nil {
		return lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Italic(true).
			Render("Completing journey...")
	}

	if m.result.IsError {
		return lipgloss.NewStyle().
			Foreground(colors.Error.Fg).
			Render(m.result.Error)
	}

	return lipgloss.NewStyle().
		Foreground(colors.Success.Fg).
		Render("Journey completed")
}

// Params returns header params.
func (m *Model) Params() []string {
	return nil
}
