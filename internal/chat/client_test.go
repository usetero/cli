package chat_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/auth/authtest"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
)

type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func TestClient_Send(t *testing.T) {
	t.Parallel()

	t.Run("sends request with correct headers", func(t *testing.T) {
		t.Parallel()

		var capturedReq *http.Request
		httpClient := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				capturedReq = req
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
					Body:       io.NopCloser(strings.NewReader("data: [DONE]\n")),
				}, nil
			},
		}

		mockAuth := &authtest.MockAuth{
			GetAccessTokenFunc: func(ctx context.Context) (string, error) {
				return "test-token", nil
			},
		}

		client := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.New(t))
		client.SetAccountID("acc-123")

		err := client.Send(context.Background(), chat.Request{
			ConversationID: "conv-1",
			Messages:       []domain.Message{},
		}, func(e chat.Event) error { return nil })

		if err != nil {
			t.Fatalf("Send() error = %v", err)
		}

		if capturedReq.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", capturedReq.Header.Get("Authorization"), "Bearer test-token")
		}
		if capturedReq.Header.Get("X-Account-ID") != "acc-123" {
			t.Errorf("X-Account-ID = %q, want %q", capturedReq.Header.Get("X-Account-ID"), "acc-123")
		}
		if capturedReq.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want %q", capturedReq.Header.Get("Content-Type"), "application/json")
		}
		if capturedReq.Header.Get("Accept") != "text/event-stream" {
			t.Errorf("Accept = %q, want %q", capturedReq.Header.Get("Accept"), "text/event-stream")
		}
	})

	t.Run("sends request to correct endpoint", func(t *testing.T) {
		t.Parallel()

		var capturedReq *http.Request
		httpClient := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				capturedReq = req
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
					Body:       io.NopCloser(strings.NewReader("data: [DONE]\n")),
				}, nil
			},
		}

		mockAuth := &authtest.MockAuth{
			GetAccessTokenFunc: func(ctx context.Context) (string, error) {
				return "token", nil
			},
		}

		client := chat.NewClientWithHTTP("https://api.example.com/", mockAuth, httpClient, logtest.New(t))

		err := client.Send(context.Background(), chat.Request{}, func(e chat.Event) error { return nil })
		if err != nil {
			t.Fatalf("Send() error = %v", err)
		}

		want := "https://api.example.com/api/chat/v1/messages"
		if capturedReq.URL.String() != want {
			t.Errorf("URL = %q, want %q", capturedReq.URL.String(), want)
		}
	})

	t.Run("returns error on auth failure", func(t *testing.T) {
		t.Parallel()

		httpClient := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				t.Fatal("HTTP client should not be called when auth fails")
				return nil, nil
			},
		}

		mockAuth := &authtest.MockAuth{
			GetAccessTokenFunc: func(ctx context.Context) (string, error) {
				return "", errors.New("token expired")
			},
		}

		client := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.New(t))

		err := client.Send(context.Background(), chat.Request{}, func(e chat.Event) error { return nil })
		if err == nil {
			t.Fatal("Send() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "access token") {
			t.Errorf("error = %q, want to contain 'access token'", err.Error())
		}
	})

	t.Run("returns error on HTTP failure", func(t *testing.T) {
		t.Parallel()

		httpClient := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("connection refused")
			},
		}

		mockAuth := &authtest.MockAuth{
			GetAccessTokenFunc: func(ctx context.Context) (string, error) {
				return "token", nil
			},
		}

		client := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.New(t))

		err := client.Send(context.Background(), chat.Request{}, func(e chat.Event) error { return nil })
		if err == nil {
			t.Fatal("Send() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "connection refused") {
			t.Errorf("error = %q, want to contain 'connection refused'", err.Error())
		}
	})

	t.Run("returns error on non-200 status", func(t *testing.T) {
		t.Parallel()

		httpClient := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(strings.NewReader("invalid token")),
				}, nil
			},
		}

		mockAuth := &authtest.MockAuth{
			GetAccessTokenFunc: func(ctx context.Context) (string, error) {
				return "token", nil
			},
		}

		client := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.New(t))

		err := client.Send(context.Background(), chat.Request{}, func(e chat.Event) error { return nil })
		if err == nil {
			t.Fatal("Send() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "401") {
			t.Errorf("error = %q, want to contain '401'", err.Error())
		}
	})

	t.Run("returns error on wrong content type", func(t *testing.T) {
		t.Parallel()

		httpClient := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(`{"error": "something"}`)),
				}, nil
			},
		}

		mockAuth := &authtest.MockAuth{
			GetAccessTokenFunc: func(ctx context.Context) (string, error) {
				return "token", nil
			},
		}

		client := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.New(t))

		err := client.Send(context.Background(), chat.Request{}, func(e chat.Event) error { return nil })
		if err == nil {
			t.Fatal("Send() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "text/event-stream") {
			t.Errorf("error = %q, want to mention expected content type", err.Error())
		}
	})

	t.Run("streams events to handler", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"type":"message_start","message_start":{"model":"claude-3"}}
data: {"type":"text_delta","text":{"content":"Hello"}}
data: {"type":"text_delta","text":{"content":" world"}}
data: {"type":"message_stop","message_stop":{"stop_reason":"end_turn"}}
data: [DONE]
`
		httpClient := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
					Body:       io.NopCloser(strings.NewReader(stream)),
				}, nil
			},
		}

		mockAuth := &authtest.MockAuth{
			GetAccessTokenFunc: func(ctx context.Context) (string, error) {
				return "token", nil
			},
		}

		client := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.New(t))

		var events []chat.Event
		err := client.Send(context.Background(), chat.Request{}, func(e chat.Event) error {
			events = append(events, e)
			return nil
		})

		if err != nil {
			t.Fatalf("Send() error = %v", err)
		}

		if len(events) != 5 {
			t.Fatalf("got %d events, want 5", len(events))
		}

		if events[0].Type != domain.BlockTypeMessageStart {
			t.Errorf("events[0].Type = %q, want %q", events[0].Type, domain.BlockTypeMessageStart)
		}
		if events[1].Type != domain.BlockTypeTextDelta {
			t.Errorf("events[1].Type = %q, want %q", events[1].Type, domain.BlockTypeTextDelta)
		}
		if events[1].Text.Content != "Hello" {
			t.Errorf("events[1].Text.Content = %q, want %q", events[1].Text.Content, "Hello")
		}
		if events[3].Type != domain.BlockTypeMessageStop {
			t.Errorf("events[3].Type = %q, want %q", events[3].Type, domain.BlockTypeMessageStop)
		}
		if !events[4].Done {
			t.Error("events[4].Done = false, want true")
		}
	})

	t.Run("stops on handler error", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"type":"text_delta","text":{"content":"Hello"}}
data: {"type":"text_delta","text":{"content":" world"}}
data: [DONE]
`
		httpClient := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
					Body:       io.NopCloser(strings.NewReader(stream)),
				}, nil
			},
		}

		mockAuth := &authtest.MockAuth{
			GetAccessTokenFunc: func(ctx context.Context) (string, error) {
				return "token", nil
			},
		}

		client := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.New(t))

		handlerErr := errors.New("handler failed")
		callCount := 0
		err := client.Send(context.Background(), chat.Request{}, func(e chat.Event) error {
			callCount++
			return handlerErr
		})

		if !errors.Is(err, handlerErr) {
			t.Errorf("Send() error = %v, want %v", err, handlerErr)
		}
		if callCount != 1 {
			t.Errorf("handler called %d times, want 1", callCount)
		}
	})
}
