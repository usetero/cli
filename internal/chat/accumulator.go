package chat

import "github.com/usetero/cli/internal/domain"

// accumulator builds a domain.Message from a stream of protocol events.
// It handles delta events by accumulating them into complete blocks.
//
// Protocol:
//   - text_delta* → content_block_stop
//   - thinking_delta* → content_block_stop
//   - tool_use → tool_input_delta* → content_block_stop
//   - message_stop (always last)
type accumulator struct {
	model      string
	stopReason string
	blocks     []domain.Block // completed blocks
	current    *domain.Block  // text/thinking block being built from deltas
	done       bool
	nextIndex  int // next block index to assign

	// Tool accumulation - tool_use starts it, deltas build input, content_block_stop finalizes
	currentTool *toolAccumulator
}

type toolAccumulator struct {
	index int
	id    string
	name  string
	input []byte
}

// newAccumulator creates a new accumulator.
func newAccumulator() *accumulator {
	return &accumulator{}
}

// handle processes a single event from the stream.
func (a *accumulator) handle(e event) {
	if e.Done {
		a.finalizeCurrent()
		a.done = true
		return
	}

	switch e.Type {
	case EventTypeMessageStart:
		if e.MessageStart != nil {
			a.model = e.MessageStart.Model
		}

	case EventTypeMessageStop:
		if e.MessageStop != nil {
			a.stopReason = e.MessageStop.StopReason
		}
		a.finalizeCurrent()

	case EventTypeTextDelta:
		a.handleTextDelta(e)

	case EventTypeThinkingDelta:
		a.handleThinkingDelta(e)

	case EventTypeToolUse:
		a.finalizeCurrent()
		if e.ToolUse == nil {
			return
		}
		// Start accumulating a new tool
		a.currentTool = &toolAccumulator{
			index: a.nextIndex,
			id:    e.ToolUse.ID,
			name:  e.ToolUse.Name,
		}
		a.nextIndex++

	case EventTypeToolInputDelta:
		// Append to current tool's input buffer
		if a.currentTool != nil {
			a.currentTool.input = append(a.currentTool.input, e.ToolInputDelta...)
		}

	case EventTypeContentBlockStop:
		// Finalize whatever block is in progress
		a.finalizeCurrent()
		a.finalizeCurrentTool()

	case EventTypeText, EventTypeThinking, EventTypeToolResult:
		// Complete blocks - just append
		a.finalizeCurrent()
		a.blocks = append(a.blocks, a.eventToBlock(e))
	}
}

// eventToBlock converts a complete block event to a domain.Block.
func (a *accumulator) eventToBlock(e event) domain.Block {
	switch e.Type {
	case EventTypeText:
		content := ""
		if e.Text != nil {
			content = e.Text.Content
		}
		return domain.Block{
			Type: domain.BlockTypeText,
			Text: &domain.TextBlock{Content: content},
		}
	case EventTypeThinking:
		content := ""
		if e.Thinking != nil {
			content = e.Thinking.Content
		}
		return domain.Block{
			Type:     domain.BlockTypeThinking,
			Thinking: &domain.Thinking{Content: content},
		}
	default:
		// Unknown type - return empty block
		return domain.Block{}
	}
}

func (a *accumulator) handleTextDelta(e event) {
	delta := ""
	if e.Text != nil {
		delta = e.Text.Content
	}

	if a.current == nil || a.current.Type != domain.BlockTypeText {
		a.finalizeCurrent()
		a.current = &domain.Block{
			Index: a.nextIndex,
			Type:  domain.BlockTypeText,
			Text:  &domain.TextBlock{Content: delta},
		}
		a.nextIndex++
	} else {
		a.current.Text.Content += delta
	}
}

func (a *accumulator) handleThinkingDelta(e event) {
	delta := ""
	if e.Thinking != nil {
		delta = e.Thinking.Content
	}

	if a.current == nil || a.current.Type != domain.BlockTypeThinking {
		a.finalizeCurrent()
		a.current = &domain.Block{
			Index:    a.nextIndex,
			Type:     domain.BlockTypeThinking,
			Thinking: &domain.Thinking{Content: delta},
		}
		a.nextIndex++
	} else {
		a.current.Thinking.Content += delta
	}
}

func (a *accumulator) finalizeCurrent() {
	if a.current != nil {
		a.blocks = append(a.blocks, *a.current)
		a.current = nil
	}
}

func (a *accumulator) finalizeCurrentTool() {
	if a.currentTool != nil {
		a.blocks = append(a.blocks, domain.Block{
			Index: a.currentTool.index,
			Type:  domain.BlockTypeToolUse,
			ToolUse: &domain.ToolUse{
				ID:            a.currentTool.id,
				Name:          a.currentTool.name,
				Input:         a.currentTool.input,
				InputComplete: true,
			},
		})
		a.currentTool = nil
	}
}

// message returns the current state of the message being built.
// This includes completed blocks plus any in-progress block.
func (a *accumulator) message() *domain.Message {
	// Build content: completed blocks + in-progress
	content := make([]domain.Block, len(a.blocks))
	copy(content, a.blocks)

	// Add any in-progress text/thinking block
	if a.current != nil {
		content = append(content, *a.current)
	}

	// Add any in-progress tool (for live rendering before content_block_stop)
	if a.currentTool != nil {
		content = append(content, domain.Block{
			Index: a.currentTool.index,
			Type:  domain.BlockTypeToolUse,
			ToolUse: &domain.ToolUse{
				ID:    a.currentTool.id,
				Name:  a.currentTool.name,
				Input: a.currentTool.input,
			},
		})
	}

	return &domain.Message{
		Role:       domain.RoleAssistant,
		Content:    content,
		Model:      a.model,
		StopReason: a.stopReason,
	}
}

// isDone returns true when the stream is complete.
func (a *accumulator) isDone() bool {
	return a.done
}
