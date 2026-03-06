package chattest

import (
	"context"

	infrachat "github.com/usetero/cli/internal/infrastructure/chat"
)

// Client is a functional mock for chat streaming client behavior.
type Client struct {
	StreamFn func(ctx context.Context, req infrachat.Request, onEvent func(infrachat.Event)) (infrachat.StreamResult, error)
}

var _ interface {
	Stream(ctx context.Context, req infrachat.Request, onEvent func(infrachat.Event)) (infrachat.StreamResult, error)
} = (*Client)(nil)

func (c Client) Stream(ctx context.Context, req infrachat.Request, onEvent func(infrachat.Event)) (infrachat.StreamResult, error) {
	if c.StreamFn == nil {
		return infrachat.StreamResult{}, nil
	}
	return c.StreamFn(ctx, req, onEvent)
}
