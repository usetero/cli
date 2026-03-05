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
)

const (
	defaultTimeout = 10 * time.Minute
	streamPath     = "/api/chat/v2/messages"
)

type TokenProvider interface {
	GetAccessToken(ctx context.Context) (string, error)
}

type Client struct {
	baseURL string
	http    *http.Client
	token   TokenProvider
}

func NewClient(baseURL string, token TokenProvider) *Client {
	return &Client{baseURL: strings.TrimSuffix(baseURL, "/"), http: &http.Client{Timeout: defaultTimeout}, token: token}
}

func NewClientWithHTTP(baseURL string, token TokenProvider, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}
	return &Client{baseURL: strings.TrimSuffix(baseURL, "/"), http: httpClient, token: token}
}

func (c *Client) Stream(ctx context.Context, req Request, onEvent func(Event)) (StreamResult, error) {
	if c == nil {
		return StreamResult{}, fmt.Errorf("chat client is nil")
	}
	if c.baseURL == "" {
		return StreamResult{}, fmt.Errorf("chat base url is required")
	}

	payload, err := req.payload()
	if err != nil {
		return StreamResult{}, err
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return StreamResult{}, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+streamPath, bytes.NewReader(body))
	if err != nil {
		return StreamResult{}, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	if c.token != nil {
		tok, err := c.token.GetAccessToken(ctx)
		if err != nil {
			return StreamResult{}, fmt.Errorf("get access token: %w", err)
		}
		if tok != "" {
			httpReq.Header.Set("Authorization", "Bearer "+tok)
		}
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return StreamResult{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return StreamResult{}, HTTPError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(b))}
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(strings.ToLower(ct), "text/event-stream") {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return StreamResult{}, fmt.Errorf("expected text/event-stream, got %q: %s", ct, strings.TrimSpace(string(b)))
	}

	var out StreamResult
	err = readSSE(resp.Body, func(e Event) error {
		if e.ConversationID != "" {
			out.ConversationID = e.ConversationID
		}
		if e.TurnID != "" {
			out.TurnID = e.TurnID
		}
		if e.Seq > out.LastSeq {
			out.LastSeq = e.Seq
		}
		if onEvent != nil {
			onEvent(e)
		}
		return nil
	})
	if err != nil {
		return StreamResult{}, err
	}
	if out.ConversationID == "" {
		out.ConversationID = req.ConversationID
	}
	return out, nil
}
