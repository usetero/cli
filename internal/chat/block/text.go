package block

// Text is prose content.
type Text struct {
	Content string `json:"content"`
}

// NewText creates a text block.
func NewText(content string) Block {
	return Block{
		Type: TypeText,
		Text: &Text{Content: content},
	}
}

// NewTextDelta creates a streaming text delta block.
func NewTextDelta(text string) Block {
	return Block{
		Type: TypeTextDelta,
		Text: &Text{Content: text},
	}
}

// EncodeText is a convenience function to encode a single text block.
func EncodeText(content string) (string, error) {
	return Encode([]Block{NewText(content)})
}
