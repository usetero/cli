package powersync

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/usetero/cli/internal/sqlite"
)

// Client handles the sync connection to the PowerSync service.
type Client struct {
	endpoint  string
	accountID string
	token     string
	db        sqlite.Database

	mu       sync.RWMutex
	status   Status
	cancel   context.CancelFunc
	doneChan chan struct{}
}

// Status represents the current state of the sync client.
type Status string

const (
	StatusDisconnected Status = "disconnected"
	StatusConnecting   Status = "connecting"
	StatusConnected    Status = "connected"
	StatusSyncing      Status = "syncing"
	StatusError        Status = "error"
)

// SyncInstruction represents an instruction from the PowerSync extension.
type SyncInstruction struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// NewClient creates a new PowerSync sync client.
func NewClient(endpoint string, db sqlite.Database) *Client {
	return &Client{
		endpoint: endpoint,
		db:       db,
		status:   StatusDisconnected,
	}
}

// Connect starts the sync connection with the given credentials.
func (c *Client) Connect(ctx context.Context, accountID, token string) error {
	c.mu.Lock()
	if c.cancel != nil {
		c.mu.Unlock()
		return fmt.Errorf("already connected")
	}

	c.accountID = accountID
	c.token = token
	c.status = StatusConnecting
	c.doneChan = make(chan struct{})

	syncCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.mu.Unlock()

	go c.syncLoop(syncCtx)

	return nil
}

// Disconnect stops the sync connection.
func (c *Client) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}

	// Wait for sync loop to finish
	if c.doneChan != nil {
		<-c.doneChan
		c.doneChan = nil
	}

	c.status = StatusDisconnected
}

// Status returns the current sync status.
func (c *Client) Status() Status {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

// setStatus updates the client status.
func (c *Client) setStatus(status Status) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status = status
}

// syncLoop runs the main sync loop.
func (c *Client) syncLoop(ctx context.Context) {
	defer close(c.doneChan)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Start sync and get instructions from extension
		result, err := c.control("start", []byte("{}"))
		if err != nil {
			c.setStatus(StatusError)
			time.Sleep(5 * time.Second)
			continue
		}

		// Process instructions from the extension
		if err := c.processInstructions(ctx, result); err != nil {
			if ctx.Err() != nil {
				return
			}
			c.setStatus(StatusError)
			time.Sleep(5 * time.Second)
			continue
		}
	}
}

// control sends a control message to the PowerSync extension.
func (c *Client) control(op string, payload []byte) ([]byte, error) {
	var result []byte
	err := c.db.QueryRow("SELECT powersync_control(?, ?)", op, payload).Scan(&result)
	if err != nil {
		return nil, fmt.Errorf("powersync_control(%s): %w", op, err)
	}
	return result, nil
}

// processInstructions processes sync instructions from the extension.
func (c *Client) processInstructions(ctx context.Context, data []byte) error {
	var instructions []SyncInstruction
	if err := json.Unmarshal(data, &instructions); err != nil {
		// Try single instruction
		var instruction SyncInstruction
		if err := json.Unmarshal(data, &instruction); err != nil {
			return fmt.Errorf("unmarshal instructions: %w", err)
		}
		instructions = []SyncInstruction{instruction}
	}

	for _, inst := range instructions {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := c.handleInstruction(ctx, inst); err != nil {
			return err
		}
	}

	return nil
}

// handleInstruction handles a single sync instruction.
func (c *Client) handleInstruction(ctx context.Context, inst SyncInstruction) error {
	switch inst.Type {
	case "EstablishSyncStream":
		return c.establishSyncStream(ctx)
	case "FetchCredentials":
		return c.sendCredentials()
	case "UploadData":
		// Write operations not yet supported
		return nil
	default:
		// Unknown instruction, ignore
		return nil
	}
}

// establishSyncStream connects to the PowerSync service and streams sync data.
func (c *Client) establishSyncStream(ctx context.Context) error {
	c.setStatus(StatusSyncing)

	// Build sync URL with account_id parameter
	syncURL, err := url.Parse(c.endpoint)
	if err != nil {
		return fmt.Errorf("parse endpoint: %w", err)
	}
	syncURL.Path = "/sync/stream"
	q := syncURL.Query()
	q.Set("account_id", c.accountID)
	syncURL.RawQuery = q.Encode()

	// Create request with auth
	req, err := http.NewRequestWithContext(ctx, "GET", syncURL.String(), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/x-ndjson")

	// Make request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connect to sync stream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("sync stream returned %d: %s", resp.StatusCode, string(body))
	}

	c.setStatus(StatusConnected)

	// Read NDJSON stream
	scanner := bufio.NewScanner(resp.Body)
	// Increase buffer size for large sync lines
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Feed line to extension via powersync_control
		result, err := c.control("line", line)
		if err != nil {
			return fmt.Errorf("process sync line: %w", err)
		}

		// Check for new instructions
		if len(result) > 0 && string(result) != "null" {
			if err := c.processInstructions(ctx, result); err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read sync stream: %w", err)
	}

	return nil
}

// sendCredentials sends the current credentials to the extension.
func (c *Client) sendCredentials() error {
	creds := map[string]string{
		"token":    c.token,
		"endpoint": c.endpoint,
	}
	credsJSON, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	_, err = c.control("credentials", credsJSON)
	if err != nil {
		return fmt.Errorf("send credentials: %w", err)
	}

	return nil
}
