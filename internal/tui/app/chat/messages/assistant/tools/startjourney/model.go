package startjourney

import (
	"encoding/json"

	"github.com/usetero/cli/internal/log"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/chat/messages/assistant/tools"
	appstartjourney "github.com/usetero/cli/internal/tui/app/tools/startjourney"
)

// Model renders and executes a start_journey tool.
type Model struct {
	theme       *styles.Theme
	logger      log.Logger
	use         *domain.ToolUse
	executor    *appstartjourney.Tool
	result      *domain.ToolResult
	journeyName string
}

// Compile-time interface check
var _ tools.Body = (*Model)(nil)

// New creates a new start_journey model.
func New(theme *styles.Theme, logger log.Logger, use *domain.ToolUse, executor *appstartjourney.Tool) *Model {
	m := &Model{
		theme:    theme,
		logger:   logger,
		use:      use,
		executor: executor,
	}

	var in struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(use.Input, &in); err == nil {
		m.journeyName = in.Name
	}

	return m
}

// Init executes the tool. Returns a Cmd that produces ResultMsg.
func (m *Model) Init() tea.Cmd {
	use := m.use
	executor := m.executor
	logger := m.logger
	journeyName := m.journeyName

	return func() tea.Msg {
		logger.Info("starting journey", "name", journeyName)

		result, err := executor.Execute(use.Input)

		if err != nil {
			logger.Error("journey start failed", "error", err)
			return tools.ResultMsg{
				ToolUseID: use.ID,
				Result: &domain.ToolResult{
					ToolUseID: use.ID,
					IsError:   true,
					Error:     err.Error(),
				},
			}
		}

		logger.Info("journey started", "name", journeyName)

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
			Render("Starting journey...")
	}

	if m.result.IsError {
		return lipgloss.NewStyle().
			Foreground(colors.Error.Fg).
			Render(m.result.Error)
	}

	return lipgloss.NewStyle().
		Foreground(colors.Success.Fg).
		Render("Journey started")
}

// Params returns header params.
func (m *Model) Params() []string {
	if m.journeyName != "" {
		return []string{m.journeyName}
	}
	return nil
}
