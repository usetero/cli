package syncertest

import (
	"context"
	"sync"

	psclient "github.com/usetero/cli/internal/infrastructure/powersync/client"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
)

type Client struct {
	mu sync.Mutex

	TokenValue   string
	SyncStreamFn func(ctx context.Context, req *psclient.SyncStreamRequest, handler psclient.LineHandler) error
}

var _ pssyncer.Client = (*Client)(nil)

func NewClient() *Client {
	return &Client{}
}

func (c *Client) SyncStream(ctx context.Context, req *psclient.SyncStreamRequest, handler psclient.LineHandler) error {
	if c.SyncStreamFn == nil {
		return nil
	}
	return c.SyncStreamFn(ctx, req, handler)
}

func (c *Client) SetToken(token psclient.AccessToken) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TokenValue = string(token)
}
