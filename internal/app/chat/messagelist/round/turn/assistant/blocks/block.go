package blocks

import "github.com/usetero/cli/internal/app/chat/messagelist/block"

// Block is the interface for all content blocks in an assistant message.
// Embeds block.Block for viewport integration and adds Index for content tracking.
type Block interface {
	block.Block

	// Index returns the block's position in the content array.
	Index() int
}
