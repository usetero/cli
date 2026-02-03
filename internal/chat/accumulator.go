package chat

import "github.com/usetero/cli/internal/domain"

// Accumulator builds blocks from a stream of events.
// It handles delta events by accumulating them into complete blocks.
type Accumulator struct {
	model      string
	stopReason string
	blocks     []domain.Block // completed blocks
	current    *domain.Block  // block being built from deltas
	done       bool
}

// NewAccumulator creates a new accumulator.
func NewAccumulator() *Accumulator {
	return &Accumulator{}
}

// Handle processes a single event from the stream.
func (a *Accumulator) Handle(event Event) {
	if event.Done {
		a.finalizeCurrent()
		a.done = true
		return
	}

	switch event.Type {
	case domain.BlockTypeMessageStart:
		if event.MessageStart != nil {
			a.model = event.MessageStart.Model
		}

	case domain.BlockTypeMessageStop:
		if event.MessageStop != nil {
			a.stopReason = event.MessageStop.StopReason
		}
		a.finalizeCurrent()

	case domain.BlockTypeTextDelta:
		a.handleTextDelta(event)

	case domain.BlockTypeThinkingDelta:
		a.handleThinkingDelta(event)

	case domain.BlockTypeToolUse:
		a.finalizeCurrent()
		a.blocks = append(a.blocks, event.Block)

	case domain.BlockTypeToolInputDelta:
		// Tool input deltas are handled by the ToolUse block itself
		// For now we ignore these - tool_use comes as a complete block

	case domain.BlockTypeText, domain.BlockTypeThinking, domain.BlockTypeToolResult:
		// Complete blocks - just append
		a.finalizeCurrent()
		a.blocks = append(a.blocks, event.Block)
	}
}

func (a *Accumulator) handleTextDelta(event Event) {
	delta := ""
	if event.Text != nil {
		delta = event.Text.Content
	}

	if a.current == nil || a.current.Type != domain.BlockTypeText {
		a.finalizeCurrent()
		a.current = &domain.Block{
			Type: domain.BlockTypeText,
			Text: &domain.TextBlock{Content: delta},
		}
	} else {
		a.current.Text.Content += delta
	}
}

func (a *Accumulator) handleThinkingDelta(event Event) {
	delta := ""
	if event.Thinking != nil {
		delta = event.Thinking.Content
	}

	if a.current == nil || a.current.Type != domain.BlockTypeThinking {
		a.finalizeCurrent()
		a.current = &domain.Block{
			Type:     domain.BlockTypeThinking,
			Thinking: &domain.Thinking{Content: delta},
		}
	} else {
		a.current.Thinking.Content += delta
	}
}

func (a *Accumulator) finalizeCurrent() {
	if a.current != nil {
		a.blocks = append(a.blocks, *a.current)
		a.current = nil
	}
}

// Blocks returns all blocks for rendering - completed plus any in-progress block.
func (a *Accumulator) Blocks() []domain.Block {
	if a.current == nil {
		return a.blocks
	}
	return append(a.blocks, *a.current)
}

// Model returns the model name from MessageStart.
func (a *Accumulator) Model() string {
	return a.model
}

// StopReason returns the stop reason from MessageStop.
func (a *Accumulator) StopReason() string {
	return a.stopReason
}

// Done returns true when the stream is complete.
func (a *Accumulator) Done() bool {
	return a.done
}

// Reset clears the accumulator for reuse.
func (a *Accumulator) Reset() {
	a.model = ""
	a.stopReason = ""
	a.blocks = nil
	a.current = nil
	a.done = false
}
