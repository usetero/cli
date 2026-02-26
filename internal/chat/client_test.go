package chat_test

import (
	"context"
	"errors"
	"fmt"
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

type blockingReadCloser struct {
	ctx context.Context
}

func (b *blockingReadCloser) Read(_ []byte) (int, error) {
	<-b.ctx.Done()
	return 0, b.ctx.Err()
}

func (b *blockingReadCloser) Close() error { return nil }

func TestClient_Stream(t *testing.T) {
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

		client := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.NewScope(t), nil)
		client.SetAccountID("acc-123")

		_, err := client.Stream(context.Background(), chat.Request{
			ConversationID: "conv-1",
			Messages:       []domain.Message{},
		}, func(msg *domain.Message) {})

		if err != nil {
			t.Fatalf("Stream() error = %v", err)
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

		client := chat.NewClientWithHTTP("https://api.example.com/", mockAuth, httpClient, logtest.NewScope(t), nil)

		_, err := client.Stream(context.Background(), chat.Request{}, func(msg *domain.Message) {})
		if err != nil {
			t.Fatalf("Stream() error = %v", err)
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

		client := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.NewScope(t), nil)

		_, err := client.Stream(context.Background(), chat.Request{}, func(msg *domain.Message) {})
		if err == nil {
			t.Fatal("Stream() expected error, got nil")
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

		client := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.NewScope(t), nil)

		_, err := client.Stream(context.Background(), chat.Request{}, func(msg *domain.Message) {})
		if err == nil {
			t.Fatal("Stream() expected error, got nil")
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

		client := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.NewScope(t), nil)

		_, err := client.Stream(context.Background(), chat.Request{}, func(msg *domain.Message) {})
		if err == nil {
			t.Fatal("Stream() expected error, got nil")
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

		client := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.NewScope(t), nil)

		_, err := client.Stream(context.Background(), chat.Request{}, func(msg *domain.Message) {})
		if err == nil {
			t.Fatal("Stream() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "text/event-stream") {
			t.Errorf("error = %q, want to mention expected content type", err.Error())
		}
	})

	t.Run("builds message from stream and calls onMessage", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"type":"message_start","message_start":{"model":"claude-3"}}
data: {"type":"text_delta","text":{"content":"Hello"}}
data: {"type":"text_delta","text":{"content":" world"}}
data: {"type":"content_block_stop"}
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

		client := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.NewScope(t), nil)

		var messages []*domain.Message
		_, err := client.Stream(context.Background(), chat.Request{}, func(msg *domain.Message) {
			// Make a copy since the message is built incrementally
			msgCopy := *msg
			messages = append(messages, &msgCopy)
		})

		if err != nil {
			t.Fatalf("Stream() error = %v", err)
		}

		// Should have received multiple updates as the message was built
		if len(messages) == 0 {
			t.Fatal("expected at least one message callback")
		}

		// Last message should have the complete content
		lastMsg := messages[len(messages)-1]
		if lastMsg.Model != "claude-3" {
			t.Errorf("Model = %q, want %q", lastMsg.Model, "claude-3")
		}
		if lastMsg.StopReason != "end_turn" {
			t.Errorf("StopReason = %q, want %q", lastMsg.StopReason, "end_turn")
		}
		if len(lastMsg.Content) != 1 {
			t.Fatalf("Content length = %d, want 1", len(lastMsg.Content))
		}
		if lastMsg.Content[0].Type != domain.BlockTypeText {
			t.Errorf("Content[0].Type = %q, want %q", lastMsg.Content[0].Type, domain.BlockTypeText)
		}
		if lastMsg.Content[0].Text.Content != "Hello world" {
			t.Errorf("Content[0].Text.Content = %q, want %q", lastMsg.Content[0].Text.Content, "Hello world")
		}
	})

	t.Run("accumulates tool use with input deltas", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"type":"message_start","message_start":{"model":"claude-3"}}
data: {"type":"tool_use","tool_use":{"id":"tool-1","name":"query"}}
data: {"type":"tool_input_delta","tool_input_delta":"{\"sql\":"}
data: {"type":"tool_input_delta","tool_input_delta":" \"SELECT 1\"}"}
data: {"type":"content_block_stop"}
data: {"type":"message_stop","message_stop":{"stop_reason":"tool_use"}}
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

		client := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.NewScope(t), nil)

		var lastMessage *domain.Message
		_, err := client.Stream(context.Background(), chat.Request{}, func(msg *domain.Message) {
			lastMessage = msg
		})

		if err != nil {
			t.Fatalf("Stream() error = %v", err)
		}

		if lastMessage == nil {
			t.Fatal("expected message")
		}
		if len(lastMessage.Content) != 1 {
			t.Fatalf("Content length = %d, want 1", len(lastMessage.Content))
		}
		if lastMessage.Content[0].Type != domain.BlockTypeToolUse {
			t.Errorf("Content[0].Type = %q, want %q", lastMessage.Content[0].Type, domain.BlockTypeToolUse)
		}
		if lastMessage.Content[0].ToolUse.ID != "tool-1" {
			t.Errorf("ToolUse.ID = %q, want %q", lastMessage.Content[0].ToolUse.ID, "tool-1")
		}
		if lastMessage.Content[0].ToolUse.Name != "query" {
			t.Errorf("ToolUse.Name = %q, want %q", lastMessage.Content[0].ToolUse.Name, "query")
		}
		expectedInput := `{"sql": "SELECT 1"}`
		if string(lastMessage.Content[0].ToolUse.Input) != expectedInput {
			t.Errorf("ToolUse.Input = %q, want %q", string(lastMessage.Content[0].ToolUse.Input), expectedInput)
		}
	})

	t.Run("WithAccountID remains stable during concurrent base account switches", func(t *testing.T) {
		t.Parallel()

		httpClient := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				if got := req.Header.Get("X-Account-ID"); got != "acc-scoped" {
					return nil, fmt.Errorf("unexpected account header %q", got)
				}
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

		base := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.NewScope(t), nil)
		base.SetAccountID("acc-base")
		scoped := base.WithAccountID("acc-scoped")

		stop := make(chan struct{})
		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				select {
				case <-stop:
					return
				default:
					base.SetAccountID("acc-a")
					base.SetAccountID("acc-b")
				}
			}
		}()

		for i := 0; i < 100; i++ {
			if _, err := scoped.Stream(context.Background(), chat.Request{}, nil); err != nil {
				close(stop)
				<-done
				t.Fatalf("scoped Stream() error at iter %d: %v", i, err)
			}
		}

		close(stop)
		<-done
	})
}

func TestClient_StreamSnapshots(t *testing.T) {
	t.Parallel()

	stream := `data: {"conversation_id":"conv-1","turn_id":"turn-1","seq":1,"type":"message_start","message_start":{"model":"claude-3"}}
data: {"conversation_id":"conv-1","turn_id":"turn-1","seq":2,"type":"text_delta","text":{"content":"Hello"}}
data: {"conversation_id":"conv-1","turn_id":"turn-1","seq":3,"type":"message_stop","message_stop":{"stop_reason":"end_turn","input_tokens":10,"output_tokens":2}}
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

	client := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.NewScope(t), nil)

	var snaps []chat.StreamSnapshot
	_, err := client.StreamSnapshots(context.Background(), chat.Request{ConversationID: "conv-1"}, func(s chat.StreamSnapshot) {
		snaps = append(snaps, s)
	})
	if err != nil {
		t.Fatalf("StreamSnapshots() error = %v", err)
	}

	if len(snaps) == 0 {
		t.Fatal("expected at least one snapshot")
	}

	last := snaps[len(snaps)-1]
	if !last.Done {
		t.Fatal("last.Done = false, want true")
	}
	if last.Status != chat.StreamStatusCompleted {
		t.Fatalf("last.Status = %q, want %q", last.Status, chat.StreamStatusCompleted)
	}
	if last.ConversationID != "conv-1" {
		t.Fatalf("last.ConversationID = %q, want conv-1", last.ConversationID)
	}
	if last.TurnID != "turn-1" {
		t.Fatalf("last.TurnID = %q, want turn-1", last.TurnID)
	}
	if last.Seq != 3 {
		t.Fatalf("last.Seq = %d, want 3", last.Seq)
	}
	if last.Metadata == nil {
		t.Fatal("last.Metadata = nil, want non-nil")
	}
}

func TestClient_StreamSnapshots_CancelledContextEmitsAbortedSnapshot(t *testing.T) {
	t.Parallel()

	httpClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
				Body:       &blockingReadCloser{ctx: req.Context()},
			}, nil
		},
	}

	mockAuth := &authtest.MockAuth{
		GetAccessTokenFunc: func(ctx context.Context) (string, error) {
			return "token", nil
		},
	}

	client := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.NewScope(t), nil)

	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(errors.New("user_cancelled"))

	var snaps []chat.StreamSnapshot
	result, err := client.StreamSnapshots(ctx, chat.Request{ConversationID: "conv-1"}, func(s chat.StreamSnapshot) {
		snaps = append(snaps, s)
	})
	if err != nil {
		t.Fatalf("StreamSnapshots() error = %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result on canceled context")
	}
	if len(snaps) == 0 {
		t.Fatal("expected aborted snapshot")
	}

	last := snaps[len(snaps)-1]
	if !last.Done {
		t.Fatal("last.Done = false, want true")
	}
	if last.Status != chat.StreamStatusAborted {
		t.Fatalf("last.Status = %q, want %q", last.Status, chat.StreamStatusAborted)
	}
	if last.AbortReason != "user_cancelled" {
		t.Fatalf("last.AbortReason = %q, want user_cancelled", last.AbortReason)
	}
}

func TestClient_StreamSnapshots_RejectsNonMonotonicSeq(t *testing.T) {
	t.Parallel()

	stream := `data: {"conversation_id":"conv-1","turn_id":"turn-1","seq":2,"type":"text_delta","text":{"content":"a"}}
data: {"conversation_id":"conv-1","turn_id":"turn-1","seq":2,"type":"text_delta","text":{"content":"b"}}
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
	client := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.NewScope(t), nil)

	_, err := client.StreamSnapshots(context.Background(), chat.Request{ConversationID: "conv-1"}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "non-monotonic seq") {
		t.Fatalf("error = %q, want non-monotonic seq", err.Error())
	}
}

func TestClient_StreamSnapshots_RejectsTurnMismatch(t *testing.T) {
	t.Parallel()

	stream := `data: {"conversation_id":"conv-1","turn_id":"turn-1","seq":1,"type":"text_delta","text":{"content":"a"}}
data: {"conversation_id":"conv-1","turn_id":"turn-2","seq":2,"type":"text_delta","text":{"content":"b"}}
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
	client := chat.NewClientWithHTTP("https://api.example.com", mockAuth, httpClient, logtest.NewScope(t), nil)

	_, err := client.StreamSnapshots(context.Background(), chat.Request{ConversationID: "conv-1"}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "turn_id mismatch") {
		t.Fatalf("error = %q, want turn_id mismatch", err.Error())
	}
}
