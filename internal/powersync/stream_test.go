package powersync_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/powersync"
)

// mockHTTPClient implements powersync.HTTPClient for testing.
type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

// mockResponse creates an http.Response with the given status and body.
func mockResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestStream_Connect(t *testing.T) {
	t.Parallel()

	t.Run("returns auth error on 401", func(t *testing.T) {
		t.Parallel()

		client := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(401, `{"error":"unauthorized"}`), nil
			},
		}

		stream := powersync.NewStreamWithClient("https://example.com", "token", client)
		err := stream.Connect(context.Background(), &powersync.StreamingSyncRequest{}, func(line []byte) error {
			return nil
		})

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !powersync.IsAuthError(err) {
			t.Errorf("expected auth error, got %v", err)
		}

		var streamErr *powersync.StreamError
		if !errors.As(err, &streamErr) {
			t.Fatalf("expected StreamError, got %T", err)
		}
		if streamErr.StatusCode != 401 {
			t.Errorf("StatusCode = %d, want 401", streamErr.StatusCode)
		}
		if streamErr.Kind != powersync.ErrorKindAuth {
			t.Errorf("Kind = %v, want ErrorKindAuth", streamErr.Kind)
		}
	})

	t.Run("returns auth error on 403", func(t *testing.T) {
		t.Parallel()

		client := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(403, `{"error":"forbidden"}`), nil
			},
		}

		stream := powersync.NewStreamWithClient("https://example.com", "token", client)
		err := stream.Connect(context.Background(), &powersync.StreamingSyncRequest{}, func(line []byte) error {
			return nil
		})

		if !powersync.IsAuthError(err) {
			t.Errorf("expected auth error, got %v", err)
		}
	})

	t.Run("returns transient error on 500", func(t *testing.T) {
		t.Parallel()

		client := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(500, `{"error":"internal server error"}`), nil
			},
		}

		stream := powersync.NewStreamWithClient("https://example.com", "token", client)
		err := stream.Connect(context.Background(), &powersync.StreamingSyncRequest{}, func(line []byte) error {
			return nil
		})

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !powersync.IsTransientError(err) {
			t.Errorf("expected transient error, got %v", err)
		}

		var streamErr *powersync.StreamError
		if errors.As(err, &streamErr) {
			if streamErr.Kind != powersync.ErrorKindTransient {
				t.Errorf("Kind = %v, want ErrorKindTransient", streamErr.Kind)
			}
		}
	})

	t.Run("returns transient error on 503", func(t *testing.T) {
		t.Parallel()

		client := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(503, `service unavailable`), nil
			},
		}

		stream := powersync.NewStreamWithClient("https://example.com", "token", client)
		err := stream.Connect(context.Background(), &powersync.StreamingSyncRequest{}, func(line []byte) error {
			return nil
		})

		if !powersync.IsTransientError(err) {
			t.Errorf("expected transient error, got %v", err)
		}
	})

	t.Run("returns transient error on 429 rate limit", func(t *testing.T) {
		t.Parallel()

		client := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(429, `rate limited`), nil
			},
		}

		stream := powersync.NewStreamWithClient("https://example.com", "token", client)
		err := stream.Connect(context.Background(), &powersync.StreamingSyncRequest{}, func(line []byte) error {
			return nil
		})

		if !powersync.IsTransientError(err) {
			t.Errorf("expected transient error for 429, got %v", err)
		}
	})

	t.Run("returns permanent error on 400", func(t *testing.T) {
		t.Parallel()

		client := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(400, `{"error":"bad request"}`), nil
			},
		}

		stream := powersync.NewStreamWithClient("https://example.com", "token", client)
		err := stream.Connect(context.Background(), &powersync.StreamingSyncRequest{}, func(line []byte) error {
			return nil
		})

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		// Should not be auth or transient
		if powersync.IsAuthError(err) {
			t.Error("400 should not be classified as auth error")
		}
		if powersync.IsTransientError(err) {
			t.Error("400 should not be classified as transient error")
		}

		var streamErr *powersync.StreamError
		if errors.As(err, &streamErr) {
			if streamErr.Kind != powersync.ErrorKindPermanent {
				t.Errorf("Kind = %v, want ErrorKindPermanent", streamErr.Kind)
			}
		}
	})

	t.Run("returns transient error on network failure", func(t *testing.T) {
		t.Parallel()

		client := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return nil, &url.Error{
					Op:  "Post",
					URL: "https://example.com",
					Err: errors.New("connection refused"),
				}
			},
		}

		stream := powersync.NewStreamWithClient("https://example.com", "token", client)
		err := stream.Connect(context.Background(), &powersync.StreamingSyncRequest{}, func(line []byte) error {
			return nil
		})

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var streamErr *powersync.StreamError
		if !errors.As(err, &streamErr) {
			t.Fatalf("expected StreamError, got %T", err)
		}
		if streamErr.Kind != powersync.ErrorKindTransient {
			t.Errorf("Kind = %v, want ErrorKindTransient", streamErr.Kind)
		}
	})

	t.Run("processes NDJSON lines on success", func(t *testing.T) {
		t.Parallel()

		responseBody := `{"type":"line1"}
{"type":"line2"}
{"type":"line3"}`

		client := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(200, responseBody), nil
			},
		}

		stream := powersync.NewStreamWithClient("https://example.com", "token", client)

		var lines []string
		err := stream.Connect(context.Background(), &powersync.StreamingSyncRequest{}, func(line []byte) error {
			lines = append(lines, string(line))
			return nil
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(lines) != 3 {
			t.Errorf("got %d lines, want 3", len(lines))
		}
	})

	t.Run("stops on handler error", func(t *testing.T) {
		t.Parallel()

		responseBody := `{"type":"line1"}
{"type":"line2"}
{"type":"line3"}`

		client := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(200, responseBody), nil
			},
		}

		stream := powersync.NewStreamWithClient("https://example.com", "token", client)

		handlerErr := errors.New("handler error")
		callCount := 0
		err := stream.Connect(context.Background(), &powersync.StreamingSyncRequest{}, func(line []byte) error {
			callCount++
			if callCount == 2 {
				return handlerErr
			}
			return nil
		})

		if !errors.Is(err, handlerErr) {
			t.Errorf("expected handler error, got %v", err)
		}
		if callCount != 2 {
			t.Errorf("handler called %d times, want 2", callCount)
		}
	})

	t.Run("sends correct authorization header", func(t *testing.T) {
		t.Parallel()

		var capturedReq *http.Request
		client := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				capturedReq = req
				return mockResponse(200, ""), nil
			},
		}

		stream := powersync.NewStreamWithClient("https://example.com", "my-token", client)
		_ = stream.Connect(context.Background(), &powersync.StreamingSyncRequest{}, func(line []byte) error {
			return nil
		})

		if capturedReq == nil {
			t.Fatal("request not captured")
		}

		auth := capturedReq.Header.Get("Authorization")
		if auth != "Bearer my-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer my-token")
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		client := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return nil, ctx.Err()
			},
		}

		stream := powersync.NewStreamWithClient("https://example.com", "token", client)
		err := stream.Connect(ctx, &powersync.StreamingSyncRequest{}, func(line []byte) error {
			return nil
		})

		if !errors.Is(err, context.Canceled) {
			// The error might be wrapped
			var streamErr *powersync.StreamError
			if errors.As(err, &streamErr) && streamErr.Err != nil {
				if !errors.Is(streamErr.Err, context.Canceled) {
					t.Errorf("expected context.Canceled, got %v", err)
				}
			}
		}
	})
}

func TestStreamError(t *testing.T) {
	t.Parallel()

	t.Run("Error() with status code", func(t *testing.T) {
		t.Parallel()

		err := &powersync.StreamError{
			Kind:       powersync.ErrorKindAuth,
			StatusCode: 401,
			Message:    "unauthorized",
		}

		got := err.Error()
		if !strings.Contains(got, "401") {
			t.Errorf("error message should contain status code: %s", got)
		}
		if !strings.Contains(got, "unauthorized") {
			t.Errorf("error message should contain message: %s", got)
		}
	})

	t.Run("Error() with wrapped error", func(t *testing.T) {
		t.Parallel()

		innerErr := errors.New("connection refused")
		err := &powersync.StreamError{
			Kind:    powersync.ErrorKindTransient,
			Message: "connection failed",
			Err:     innerErr,
		}

		got := err.Error()
		if !strings.Contains(got, "connection refused") {
			t.Errorf("error message should contain wrapped error: %s", got)
		}
	})

	t.Run("Unwrap() returns inner error", func(t *testing.T) {
		t.Parallel()

		innerErr := errors.New("inner")
		err := &powersync.StreamError{Err: innerErr}

		if err.Unwrap() != innerErr {
			t.Error("Unwrap() should return inner error")
		}
	})
}
