package usecase

import (
	"context"

	corechat "github.com/usetero/cli/internal/core/chat"
)

// ChatGateway is the use-case boundary for chat streaming.
type ChatGateway interface {
	StreamSnapshots(ctx context.Context, req StreamRequest, onSnapshot func(corechat.StreamSnapshot)) (*corechat.StreamResult, error)
}
