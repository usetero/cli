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

// HTTPClient is the interface for making HTTP requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Streamer is the interface for connecting to the PowerSync sync service.
type Streamer interface {
	Connect(ctx context.Context, req *StreamingSyncRequest, handler LineHandler) error
	SetToken(token string)
}

// LineHandler is called for each line received from the sync stream.
type LineHandler func(line []byte) error

// Stream handles the HTTP connection to the PowerSync sync service.
type Stream struct {
	endpoint string
	token    string
	client   HTTPClient
}

var _ Streamer = (*Stream)(nil)

// NewStream creates a new sync stream client.
func NewStream(endpoint, token string) *Stream {
	return &Stream{
		endpoint: endpoint,
		token:    token,
		client:   http.DefaultClient,
	}
}

// NewStreamWithClient creates a new sync stream client with a custom HTTP client.
func NewStreamWithClient(endpoint, token string, client HTTPClient) *Stream {
	return &Stream{
		endpoint: endpoint,
		token:    token,
		client:   client,
	}
}

// SetToken updates the authentication token.
func (s *Stream) SetToken(token string) {
	s.token = token
}

// Connect opens a sync stream and calls handler for each received line.
// It blocks until the stream ends, context is cancelled, or handler returns an error.
func (s *Stream) Connect(ctx context.Context, req *StreamingSyncRequest, handler LineHandler) error {
	syncURL, err := url.Parse(s.endpoint)
	if err != nil {
		return fmt.Errorf("parse endpoint: %w", err)
	}
	syncURL.Path = "/sync/stream"

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, syncURL.String(), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/x-ndjson")
	httpReq.Body = io.NopCloser(bytes.NewReader(body))
	httpReq.ContentLength = int64(len(body))

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return &StreamError{
			Kind:    StreamErrorTransient,
			Message: "connection failed",
			Err:     err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return &StreamError{
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

// StreamingSyncRequest is the request body for the sync stream endpoint.
type StreamingSyncRequest struct {
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

// StreamErrorKind classifies stream errors.
type StreamErrorKind int

const (
	StreamErrorTransient StreamErrorKind = iota // Retry with backoff (5xx, 429, network errors)
	StreamErrorAuth                             // Refresh token and retry (401, 403)
	StreamErrorPermanent                        // Don't retry (4xx)
)

// StreamError represents an error from the sync stream.
type StreamError struct {
	Kind       StreamErrorKind
	StatusCode int
	Message    string
	Err        error
}

func (e *StreamError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("stream error %d: %s", e.StatusCode, e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("stream error: %v", e.Err)
	}
	return fmt.Sprintf("stream error: %s", e.Message)
}

func (e *StreamError) Unwrap() error {
	return e.Err
}

func (e *StreamError) IsAuth() bool {
	return e.Kind == StreamErrorAuth
}

func (e *StreamError) IsTransient() bool {
	return e.Kind == StreamErrorTransient
}

func classifyHTTPStatus(statusCode int) StreamErrorKind {
	switch {
	case statusCode == 401 || statusCode == 403:
		return StreamErrorAuth
	case statusCode >= 500 || statusCode == 429:
		return StreamErrorTransient
	default:
		return StreamErrorPermanent
	}
}
