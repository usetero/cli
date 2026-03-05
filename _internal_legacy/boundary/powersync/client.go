// Package api provides HTTP client access to the PowerSync service.
package powersync

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client provides HTTP access to the PowerSync service.
type Client interface {
	SyncStream(ctx context.Context, req *SyncStreamRequest, handler LineHandler) error
	GetWriteCheckpoint(ctx context.Context, clientID string) (string, error)
	SetToken(token string)
}

// HTTPDoer is the interface for making HTTP requests.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewClient creates a new PowerSync client.
func NewClient(endpoint string) Client {
	return &httpClient{
		endpoint: endpoint,
		http:     http.DefaultClient,
	}
}

// httpClient is the concrete implementation of Client.
type httpClient struct {
	endpoint string
	token    string
	http     HTTPDoer
}

// SetToken updates the authentication token.
func (c *httpClient) SetToken(token string) {
	c.token = token
}

// SetHTTPDoer sets a custom HTTP client (for testing).
func (c *httpClient) SetHTTPDoer(doer HTTPDoer) {
	c.http = doer
}

// LineHandler is called for each line received from the sync stream.
type LineHandler func(line []byte) error

// SyncStream opens a streaming connection to the sync endpoint.
// It calls handler for each line received until the stream ends,
// context is cancelled, or handler returns an error.
func (c *httpClient) SyncStream(ctx context.Context, req *SyncStreamRequest, handler LineHandler) error {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return fmt.Errorf("parse endpoint: %w", err)
	}
	u.Path = "/sync/stream"

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/x-ndjson")
	httpReq.Body = io.NopCloser(bytes.NewReader(body))
	httpReq.ContentLength = int64(len(body))

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return &Error{
			Kind:    ErrorKindTransient,
			Message: "connection failed",
			Err:     err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return &Error{
			Kind:       classifyHTTPStatus(resp.StatusCode),
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)

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

		if err := handler(line); err != nil {
			return err
		}
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read stream: %w", err)
	}

	return nil
}

// GetWriteCheckpoint fetches a write checkpoint from the server.
// The checkpoint represents the server's acknowledgment of uploaded changes.
func (c *httpClient) GetWriteCheckpoint(ctx context.Context, clientID string) (string, error) {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return "", fmt.Errorf("parse endpoint: %w", err)
	}
	u.Path = "/write-checkpoint2.json"
	u.RawQuery = "client_id=" + url.QueryEscape(clientID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		WriteCheckpoint string `json:"write_checkpoint"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	return result.WriteCheckpoint, nil
}

// SyncStreamRequest is the request body for the sync stream endpoint.
type SyncStreamRequest struct {
	Buckets         []BucketRequest     `json:"buckets"`
	IncludeChecksum bool                `json:"include_checksum"`
	RawData         bool                `json:"raw_data"`
	BinaryData      bool                `json:"binary_data"`
	ClientID        string              `json:"client_id"`
	Parameters      map[string]any      `json:"parameters,omitempty"`
	Streams         *StreamSubscription `json:"streams,omitempty"`
	AppMetadata     json.RawMessage     `json:"app_metadata,omitempty"`
}

// BucketRequest specifies a bucket to sync and the last known checkpoint.
type BucketRequest struct {
	Name  string `json:"name"`
	After string `json:"after"`
}

// StreamSubscription defines stream subscription preferences.
type StreamSubscription struct {
	IncludeDefaults bool                          `json:"include_defaults"`
	Subscriptions   []RequestedStreamSubscription `json:"subscriptions"`
}

// RequestedStreamSubscription is a request to subscribe to a stream.
type RequestedStreamSubscription struct {
	Stream           string `json:"stream"`
	Parameters       string `json:"parameters,omitempty"`
	OverridePriority *int   `json:"override_priority,omitempty"`
}

// ErrorKind classifies client errors.
type ErrorKind int

const (
	ErrorKindTransient ErrorKind = iota // Retry with backoff
	ErrorKindAuth                       // Refresh token and retry
	ErrorKindPermanent                  // Don't retry
)

// Error represents an error from the PowerSync service.
type Error struct {
	Kind       ErrorKind
	StatusCode int
	Message    string
	Err        error
}

func (e *Error) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("powersync api: %d: %s", e.StatusCode, e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("powersync api: %v", e.Err)
	}
	return fmt.Sprintf("powersync api: %s", e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) IsAuth() bool {
	return e.Kind == ErrorKindAuth
}

func (e *Error) IsTransient() bool {
	return e.Kind == ErrorKindTransient
}

func (e *Error) IsPermanent() bool {
	return e.Kind == ErrorKindPermanent
}

func classifyHTTPStatus(statusCode int) ErrorKind {
	switch {
	case statusCode == 401 || statusCode == 403:
		return ErrorKindAuth
	case statusCode >= 500 || statusCode == 429:
		return ErrorKindTransient
	default:
		return ErrorKindPermanent
	}
}
