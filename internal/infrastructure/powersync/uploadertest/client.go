package uploadertest

import (
	"context"
	"sync"
	"sync/atomic"

	psclient "github.com/usetero/cli/internal/infrastructure/powersync/client"
)

type Client struct {
	mu sync.Mutex

	TokenValue      psclient.AccessToken
	WriteCheckpoint psclient.WriteCheckpoint
	TokenSetCalls   atomic.Int32
}

func (c *Client) SyncStream(context.Context, *psclient.SyncStreamRequest, psclient.LineHandler) error {
	return nil
}

func (c *Client) SetToken(token psclient.AccessToken) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TokenValue = token
	c.TokenSetCalls.Add(1)
}

func (c *Client) GetWriteCheckpoint(context.Context, psclient.ClientID) (psclient.WriteCheckpoint, error) {
	return c.WriteCheckpoint, nil
}
