package block

import "encoding/json"

// Accumulator accumulates streaming content blocks into a structured array.
// It merges consecutive deltas of the same type into single blocks.
type Accumulator struct {
	blocks []Block
}

// NewAccumulator creates a new content block accumulator.
func NewAccumulator() *Accumulator {
	return &Accumulator{}
}

// Apply processes a block and returns true if the accumulator state changed.
// Handles delta merging automatically - text and thinking deltas are appended
// to existing blocks of the same type, tool input deltas are appended to the
// current tool's RawInput, and other blocks are added as new entries.
func (a *Accumulator) Apply(b Block) bool {
	switch b.Type {
	case TypeTextDelta:
		if b.Text != nil {
			a.AppendText(b.Text.Content)
			return true
		}
	case TypeThinkingDelta:
		if b.Thinking != nil {
			a.AppendThinking(b.Thinking.Content)
			return true
		}
	case TypeToolUse:
		if b.ToolUse != nil {
			a.AddToolUse(b.ToolUse)
			return true
		}
	case TypeToolInputDelta:
		if b.ToolInputDelta != "" {
			return a.AppendToolInput(b.ToolInputDelta)
		}
	case TypeToolResult:
		if b.ToolResult != nil {
			a.AddToolResult(b.ToolResult)
			return true
		}
	}
	return false
}

// AppendText appends text to the last text block, or creates a new one.
func (a *Accumulator) AppendText(text string) {
	// Append to last block if it's a text block
	if n := len(a.blocks); n > 0 {
		last := &a.blocks[n-1]
		if last.Type == TypeText && last.Text != nil {
			last.Text.Content += text
			return
		}
	}

	// Create new text block
	a.blocks = append(a.blocks, Block{
		Type: TypeText,
		Text: &Text{Content: text},
	})
}

// AppendThinking appends thinking to the last thinking block, or creates a new one.
func (a *Accumulator) AppendThinking(thinking string) {
	// Append to last block if it's a thinking block
	if n := len(a.blocks); n > 0 {
		last := &a.blocks[n-1]
		if last.Type == TypeThinking && last.Thinking != nil {
			last.Thinking.Content += thinking
			return
		}
	}

	// Create new thinking block
	a.blocks = append(a.blocks, Block{
		Type:     TypeThinking,
		Thinking: &Thinking{Content: thinking},
	})
}

// AddToolUse adds a tool use block.
func (a *Accumulator) AddToolUse(tool *ToolUse) {
	a.blocks = append(a.blocks, Block{
		Type:    TypeToolUse,
		ToolUse: tool,
	})
}

// AppendToolInput appends input JSON to the most recent tool use block.
// Returns true if input was appended, false if no tool block exists.
func (a *Accumulator) AppendToolInput(input string) bool {
	// Find the last tool use block
	for i := len(a.blocks) - 1; i >= 0; i-- {
		if a.blocks[i].Type == TypeToolUse && a.blocks[i].ToolUse != nil {
			a.blocks[i].ToolUse.RawInput += input
			return true
		}
	}
	// No tool block found - ignore (shouldn't happen with valid stream)
	return false
}

// AddToolResult adds a tool result block.
func (a *Accumulator) AddToolResult(result *ToolResult) {
	a.blocks = append(a.blocks, Block{
		Type:       TypeToolResult,
		ToolResult: result,
	})
}

// JSON returns the blocks as a JSON string.
func (a *Accumulator) JSON() string {
	if len(a.blocks) == 0 {
		return "[]"
	}
	data, err := json.Marshal(a.blocks)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// Blocks returns the accumulated blocks.
func (a *Accumulator) Blocks() []Block {
	return a.blocks
}
