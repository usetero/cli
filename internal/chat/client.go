// Package chat provides a client for the stateless Chat API.
//
// The Chat API is a pure function: f(messages, context) → stream of blocks.
// It doesn't read or write messages to any database. The client sends the
// full conversation history on every request and receives a streamed response.
//
// Message persistence is handled separately by the caller.
package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
)

const (
	retryMax     = 3
	retryWaitMin = 100 * time.Millisecond
	retryWaitMax = 2 * time.Second
)

// HTTPDoer is the interface for making HTTP requests.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client sends messages to the Chat API and streams responses.
type Client interface {
	// Stream sends the conversation to the Chat API and streams the response.
	// The onMessage callback is called each time the message is updated with new content.
	// The returned *domain.Message grows as content streams in.
	Stream(ctx context.Context, req Request, onMessage func(*domain.Message)) error

	// SetAccountID sets the account ID for requests.
	SetAccountID(accountID domain.AccountID)
}

// client is the concrete implementation of Client.
type client struct {
	endpoint    string
	httpClient  HTTPDoer
	auth        auth.Auth
	accountID   domain.AccountID
	logger      log.Logger
	globalTools []Tool
}

// Ensure client implements Client.
var _ Client = (*client)(nil)

// NewClient creates a new Chat API client.
// - globalTools are included in every request automatically
// - Retries transient errors (connection reset, 502/503/504) up to 3 times with backoff
// - Gets a fresh token via auth.GetAccessToken before each request
func NewClient(endpoint string, authService auth.Auth, logger log.Logger, globalTools []Tool) Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = retryMax
	retryClient.RetryWaitMin = retryWaitMin
	retryClient.RetryWaitMax = retryWaitMax
	retryClient.Logger = nil

	return &client{
		endpoint:    strings.TrimSuffix(endpoint, "/"),
		httpClient:  retryClient.StandardClient(),
		auth:        authService,
		logger:      logger,
		globalTools: globalTools,
	}
}

// NewClientWithHTTP creates a new Chat API client with a custom HTTP client (for testing).
func NewClientWithHTTP(endpoint string, authService auth.Auth, httpClient HTTPDoer, logger log.Logger, globalTools []Tool) Client {
	return &client{
		endpoint:    strings.TrimSuffix(endpoint, "/"),
		httpClient:  httpClient,
		auth:        authService,
		logger:      logger,
		globalTools: globalTools,
	}
}

// SetAccountID sets the account ID for requests.
func (c *client) SetAccountID(accountID domain.AccountID) {
	c.accountID = accountID
}

// Stream sends the conversation to the Chat API and streams the response.
// The onMessage callback is called each time the message is updated with new content.
// Global tools are automatically merged with any request-specific tools.
func (c *client) Stream(ctx context.Context, req Request, onMessage func(*domain.Message)) error {
	// Merge global tools with request-specific tools
	allTools := append(c.globalTools, req.Tools...)
	req.Tools = allTools

	c.logger.Debug("sending to chat API",
		log.String("conversation_id", req.ConversationID),
		log.Int("message_count", len(req.Messages)),
		log.Int("context_count", len(req.Context)),
		log.Int("tool_count", len(req.Tools)),
	)

	// Get fresh token for this request
	token, err := c.auth.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("get access token: %w", err)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/api/chat/v1/messages", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	c.setHeaders(httpReq, token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chat API error %d: %s", resp.StatusCode, string(body))
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/event-stream") {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expected text/event-stream, got %s: %s", contentType, string(body))
	}

	// Use internal accumulator to build the message from events
	acc := newAccumulator()

	return readStream(resp.Body, func(e event) error {
		acc.handle(e)
		if onMessage != nil {
			onMessage(acc.message())
		}
		return nil
	})
}

func (c *client) setHeaders(req *http.Request, token string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer "+token)
	if c.accountID != "" {
		req.Header.Set("X-Account-ID", c.accountID.String())
	}
}
