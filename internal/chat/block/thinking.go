package block

// Thinking is the AI's internal reasoning.
type Thinking struct {
	Content string `json:"content"`
}

// NewThinking creates a thinking block.
func NewThinking(content string) Block {
	return Block{
		Type:     TypeThinking,
		Thinking: &Thinking{Content: content},
	}
}

// NewThinkingDelta creates a streaming thinking delta block.
func NewThinkingDelta(thinking string) Block {
	return Block{
		Type:     TypeThinkingDelta,
		Thinking: &Thinking{Content: thinking},
	}
}
