package powersync

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// StreamError represents an error from the sync stream with classification.
type StreamError struct {
	Kind       ErrorKind
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

// ErrorKind classifies stream errors for appropriate handling.
type ErrorKind int

const (
	// ErrorKindTransient represents temporary failures that should be retried with backoff.
	// Examples: 500, 502, 503, 504, network timeouts, connection refused.
	ErrorKindTransient ErrorKind = iota

	// ErrorKindAuth represents authentication/authorization failures.
	// Examples: 401, 403. Should trigger token refresh and immediate retry.
	ErrorKindAuth

	// ErrorKindPermanent represents errors that won't resolve with retries.
	// Examples: 400 (bad request), invalid configuration.
	ErrorKindPermanent
)

// ClassifyHTTPError determines the error kind based on HTTP status code.
func ClassifyHTTPError(statusCode int) ErrorKind {
	switch {
	case statusCode == 401 || statusCode == 403:
		return ErrorKindAuth
	case statusCode >= 500 || statusCode == 429:
		return ErrorKindTransient
	default:
		// 4xx errors (except 401, 403, 429) are permanent
		return ErrorKindPermanent
	}
}

// IsAuthError returns true if the error is an authentication error.
func IsAuthError(err error) bool {
	var streamErr *StreamError
	if errors.As(err, &streamErr) {
		return streamErr.Kind == ErrorKindAuth
	}
	return false
}

// IsTransientError returns true if the error is transient and should be retried.
func IsTransientError(err error) bool {
	var streamErr *StreamError
	if errors.As(err, &streamErr) {
		return streamErr.Kind == ErrorKindTransient
	}
	// Network errors are transient
	return isNetworkError(err)
}

// isNetworkError checks if an error is a network-related error.
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	// Check for common network error patterns
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return urlErr.Timeout() || urlErr.Temporary()
	}
	return false
}

// HTTPClient is the interface for making HTTP requests.
// *http.Client satisfies this interface.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Streamer is the interface for connecting to the PowerSync sync service.
// This allows mocking the stream in tests.
type Streamer interface {
	Connect(ctx context.Context, req *StreamingSyncRequest, handler LineHandler) error
	SetToken(token string)
}

// Ensure Stream implements Streamer.
var _ Streamer = (*Stream)(nil)

// Stream handles the HTTP connection to the PowerSync sync service.
type Stream struct {
	endpoint string
	token    string
	client   HTTPClient
}

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

// LineHandler is called for each line received from the sync stream.
// Return an error to stop processing.
type LineHandler func(line []byte) error

// Connect opens a sync stream and calls handler for each received line.
// It blocks until the stream ends, context is cancelled, or handler returns an error.
func (s *Stream) Connect(ctx context.Context, req *StreamingSyncRequest, handler LineHandler) error {
	syncURL, err := url.Parse(s.endpoint)
	if err != nil {
		return fmt.Errorf("parse endpoint: %w", err)
	}
	syncURL.Path = "/sync/stream"

	// Marshal request body
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	// Create POST request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, syncURL.String(), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/x-ndjson")
	httpReq.Body = io.NopCloser(bytes.NewReader(body))
	httpReq.ContentLength = int64(len(body))

	// Make request
	resp, err := s.client.Do(httpReq)
	if err != nil {
		// Classify network errors as transient
		kind := ErrorKindTransient
		if isNetworkError(err) {
			kind = ErrorKindTransient
		}
		return &StreamError{
			Kind:    kind,
			Message: "connection failed",
			Err:     err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return &StreamError{
			Kind:       ClassifyHTTPError(resp.StatusCode),
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	// Read NDJSON stream
	scanner := bufio.NewScanner(resp.Body)
	// Increase buffer for large sync payloads (up to 16MB)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)

	for scanner.Scan() {
		// Check context between lines - scanner.Scan() blocks on I/O,
		// but context cancellation closes the connection which unblocks it.
		// This check catches cancellation that happened during processing.
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

	// Check if we exited due to context cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read stream: %w", err)
	}

	return nil
}
