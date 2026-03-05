package chat

import "github.com/usetero/cli/internal/domain"

// StreamResult captures the final stream outputs for a turn.
type StreamResult struct {
	Message  *domain.Message
	Metadata *StreamMetadata // nil if no metadata_update event was received
}
