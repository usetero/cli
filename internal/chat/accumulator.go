package chat

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/usetero/cli/internal/domain"
)

// accumulator builds a domain.Message from a stream of protocol events.
type accumulator struct {
	model         string
	stopReason    string
	blocks        []domain.Block
	current       *domain.Block
	nextIndex     int
	title         string
	contextWindow int
	inputTokens   int
	outputTokens  int

	openTools     map[string]*toolAccumulator
	openToolOrder []string
	seenTools     map[string]struct{}
}

type toolAccumulator struct {
	index int
	id    string
	name  string
	input []byte
}

func newAccumulator() *accumulator {
	return &accumulator{
		openTools: make(map[string]*toolAccumulator),
		seenTools: make(map[string]struct{}),
	}
}

func (a *accumulator) handle(e event) error {
	if e.Done {
		return nil
	}

	switch e.Type {
	case EventTypeMessageStart:
		a.model = e.MessageStart.Model
		a.contextWindow = *e.MessageStart.ContextWindow
		return nil

	case EventTypeMessageStop:
		if len(a.openTools) > 0 {
			return fmt.Errorf("protocol error: message_stop with %d unfinished tool blocks", len(a.openTools))
		}
		a.stopReason = e.MessageStop.StopReason
		a.inputTokens = *e.MessageStop.InputTokens
		a.outputTokens = *e.MessageStop.OutputTokens
		a.finalizeCurrent()
		return nil

	case EventTypeTextDelta:
		a.handleTextDelta(*e.Text.Content)
		return nil

	case EventTypeThinkingDelta:
		a.handleThinkingDelta(*e.Thinking.Content)
		return nil

	case EventTypeToolUse:
		a.finalizeCurrent()
		if _, exists := a.seenTools[e.ToolUse.ID]; exists {
			return fmt.Errorf("protocol error: duplicate tool_use id %q", e.ToolUse.ID)
		}
		a.seenTools[e.ToolUse.ID] = struct{}{}
		a.openTools[e.ToolUse.ID] = &toolAccumulator{
			index: a.nextIndex,
			id:    e.ToolUse.ID,
			name:  e.ToolUse.Name,
		}
		a.openToolOrder = append(a.openToolOrder, e.ToolUse.ID)
		a.nextIndex++
		return nil

	case EventTypeToolInputDelta:
		tool, ok := a.openTools[e.ToolUseID]
		if !ok {
			return fmt.Errorf("protocol error: tool_input_delta for unknown tool_use_id %q", e.ToolUseID)
		}
		tool.input = append(tool.input, e.ToolInputDelta...)
		return nil

	case EventTypeContentBlockStop:
		if e.ToolUseID == "" {
			if len(a.openTools) > 0 && a.current == nil {
				return fmt.Errorf("protocol error: content_block_stop missing tool_use_id for open tool block")
			}
			a.finalizeCurrent()
			return nil
		}
		return a.finalizeTool(e.ToolUseID)

	case EventTypeMetadataUpdate:
		if e.Metadata.Title != "" {
			a.title = e.Metadata.Title
		}
		return nil
	}

	return fmt.Errorf("protocol error: unhandled event type %q", e.Type)
}

func (a *accumulator) handleTextDelta(delta string) {
	if a.current == nil || a.current.Type != domain.BlockTypeText {
		a.finalizeCurrent()
		a.current = &domain.Block{
			Index: a.nextIndex,
			Type:  domain.BlockTypeText,
			Text:  &domain.TextBlock{Content: delta},
		}
		a.nextIndex++
		return
	}
	a.current.Text.Content += delta
}

func (a *accumulator) handleThinkingDelta(delta string) {
	if a.current == nil || a.current.Type != domain.BlockTypeThinking {
		a.finalizeCurrent()
		a.current = &domain.Block{
			Index:    a.nextIndex,
			Type:     domain.BlockTypeThinking,
			Thinking: &domain.Thinking{Content: delta},
		}
		a.nextIndex++
		return
	}
	a.current.Thinking.Content += delta
}

func (a *accumulator) finalizeCurrent() {
	if a.current == nil {
		return
	}
	a.blocks = append(a.blocks, *a.current)
	a.current = nil
}

func (a *accumulator) finalizeTool(toolUseID string) error {
	tool, ok := a.openTools[toolUseID]
	if !ok {
		return fmt.Errorf("protocol error: content_block_stop for unknown tool_use_id %q", toolUseID)
	}
	if !json.Valid(tool.input) {
		return fmt.Errorf("protocol error: invalid JSON tool input for %q", toolUseID)
	}

	a.blocks = append(a.blocks, domain.Block{
		Index: tool.index,
		Type:  domain.BlockTypeToolUse,
		ToolUse: &domain.ToolUse{
			ID:            tool.id,
			Name:          tool.name,
			Input:         json.RawMessage(append([]byte(nil), tool.input...)),
			InputComplete: true,
		},
	})
	delete(a.openTools, toolUseID)
	return nil
}

func (a *accumulator) message() *domain.Message {
	content := make([]domain.Block, len(a.blocks))
	copy(content, a.blocks)

	if a.current != nil {
		content = append(content, *a.current)
	}

	for _, id := range a.openToolOrder {
		tool, ok := a.openTools[id]
		if !ok {
			continue
		}
		content = append(content, domain.Block{
			Index: tool.index,
			Type:  domain.BlockTypeToolUse,
			ToolUse: &domain.ToolUse{
				ID:    tool.id,
				Name:  tool.name,
				Input: json.RawMessage(append([]byte(nil), tool.input...)),
			},
		})
	}

	sort.Slice(content, func(i, j int) bool {
		return content[i].Index < content[j].Index
	})

	return &domain.Message{
		Role:       domain.RoleAssistant,
		Content:    content,
		Model:      a.model,
		StopReason: a.stopReason,
	}
}
