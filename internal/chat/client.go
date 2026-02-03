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
// This interface allows services to be tested without real API calls.
type Client interface {
	Send(ctx context.Context, req Request, handler Handler) error
	SetAccountID(accountID string)
}

// client is the concrete implementation of Client.
type client struct {
	endpoint   string
	httpClient HTTPDoer
	auth       auth.Auth
	accountID  string
	logger     log.Logger
}

// Ensure client implements Client.
var _ Client = (*client)(nil)

// NewClient creates a new Chat API client.
// - Retries transient errors (connection reset, 502/503/504) up to 3 times with backoff
// - Gets a fresh token via auth.GetAccessToken before each request
func NewClient(endpoint string, authService auth.Auth, logger log.Logger) Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = retryMax
	retryClient.RetryWaitMin = retryWaitMin
	retryClient.RetryWaitMax = retryWaitMax
	retryClient.Logger = nil

	return &client{
		endpoint:   strings.TrimSuffix(endpoint, "/"),
		httpClient: retryClient.StandardClient(),
		auth:       authService,
		logger:     logger,
	}
}

// NewClientWithHTTP creates a new Chat API client with a custom HTTP client (for testing).
func NewClientWithHTTP(endpoint string, authService auth.Auth, httpClient HTTPDoer, logger log.Logger) Client {
	return &client{
		endpoint:   strings.TrimSuffix(endpoint, "/"),
		httpClient: httpClient,
		auth:       authService,
		logger:     logger,
	}
}

// SetAccountID sets the account ID for requests.
func (c *client) SetAccountID(accountID string) {
	c.accountID = accountID
}

// Handler is called for each event in the response stream.
type Handler func(Event) error

// Send sends the conversation to the Chat API and streams the response.
//
// The handler is called for each event:
//   - MessageStart: contains model info
//   - TextDelta, ThinkingDelta, ToolInputDelta: streaming content
//   - ToolUse: complete tool call
//   - MessageStop: contains stop_reason, signals end of response
//   - Done: stream complete (after MessageStop)
func (c *client) Send(ctx context.Context, req Request, handler Handler) error {
	c.logger.Debug("sending to chat API",
		log.String("conversation_id", req.ConversationID),
		log.Int("message_count", len(req.Messages)),
		log.Int("context_count", len(req.Context)),
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

	return readStream(resp.Body, handler)
}

func (c *client) setHeaders(req *http.Request, token string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer "+token)
	if c.accountID != "" {
		req.Header.Set("X-Account-ID", c.accountID)
	}
}
