package chat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client communicates with the Tero Chat API.
type Client struct {
	endpoint   string
	httpClient *http.Client
	token      string
	accountID  string
}

// NewClient creates a new Chat API client.
func NewClient(endpoint string) *Client {
	return &Client{
		endpoint:   strings.TrimSuffix(endpoint, "/"),
		httpClient: &http.Client{},
	}
}

// SetToken sets the authentication token.
func (c *Client) SetToken(token string) {
	c.token = token
}

// SetAccountID sets the account ID for requests.
func (c *Client) SetAccountID(accountID string) {
	c.accountID = accountID
}

// MessageRole is the role of a message sender.
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
)

// ContentBlock represents a block of content in a message.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// SendMessageRequest is the request body for sending a message.
type SendMessageRequest struct {
	MessageID      string         `json:"message_id"`
	ConversationID string         `json:"conversation_id"`
	Role           MessageRole    `json:"role"`
	Content        []ContentBlock `json:"content"`
	// Required for assistant messages only
	Model      string `json:"model,omitempty"`
	StopReason string `json:"stop_reason,omitempty"`
}

// SendMessageResponse is the response for assistant message saves.
type SendMessageResponse struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	CreatedAt      string `json:"created_at"`
}

// StreamEvent represents an SSE event from the chat stream.
type StreamEvent struct {
	Type string          `json:"type"`
	Text *TextDelta      `json:"text,omitempty"`
	Tool *ToolUseEvent   `json:"tool_use,omitempty"`
	Done bool            `json:"-"` // Set when we receive [DONE]
	Raw  json.RawMessage `json:"-"` // The raw event data
}

// TextDelta contains streamed text content.
type TextDelta struct {
	Content string `json:"content"`
}

// ToolUseEvent contains tool use information.
type ToolUseEvent struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// StreamHandler is called for each event in the stream.
type StreamHandler func(event StreamEvent) error

// SendUserMessage sends a user message and streams the response.
// The handler is called for each SSE event received.
func (c *Client) SendUserMessage(ctx context.Context, req SendMessageRequest, handler StreamHandler) error {
	if req.Role != RoleUser {
		return fmt.Errorf("SendUserMessage requires role=user")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/api/chat/v1/messages", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	c.setHeaders(httpReq)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chat API error %d: %s", resp.StatusCode, string(body))
	}

	return c.readSSEStream(resp.Body, handler)
}

// SaveAssistantMessage saves an assistant message for durability.
func (c *Client) SaveAssistantMessage(ctx context.Context, req SendMessageRequest) (*SendMessageResponse, error) {
	if req.Role != RoleAssistant {
		return nil, fmt.Errorf("SaveAssistantMessage requires role=assistant")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/api/chat/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("chat API error %d: %s", resp.StatusCode, string(body))
	}

	var result SendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if c.accountID != "" {
		req.Header.Set("X-Account-ID", c.accountID)
	}
}

func (c *Client) readSSEStream(r io.Reader, handler StreamHandler) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Parse SSE data line
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// Check for stream end
		if data == "[DONE]" {
			return handler(StreamEvent{Done: true})
		}

		// Parse JSON event
		var event StreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return fmt.Errorf("parse SSE event: %w", err)
		}
		event.Raw = json.RawMessage(data)

		if err := handler(event); err != nil {
			return err
		}
	}

	return scanner.Err()
}
