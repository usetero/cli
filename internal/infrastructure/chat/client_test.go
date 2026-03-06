package chat

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type staticToken struct{ token string }

func (s staticToken) GetAccessToken(context.Context) (string, error) { return s.token, nil }

func TestClientStream(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method: %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Fatalf("auth header: %q", got)
		}
		b, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(b), "chat_protocol_version") {
			t.Fatalf("missing protocol version in body")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"chat_stream_version\":\"v2\",\"conversation_id\":\"e7fdf7ec-fce5-4ca6-a572-bfd6bf8df3c8\",\"turn_id\":\"turn-1\",\"seq\":1,\"type\":\"text_delta\",\"text\":{\"content\":\"hello\"}}\n")
		_, _ = io.WriteString(w, "data: [DONE]\n")
	}))
	defer ts.Close()

	client := NewClient(ts.URL, staticToken{token: "tok"})
	res, err := client.Stream(context.Background(), Request{
		ConversationID: "e7fdf7ec-fce5-4ca6-a572-bfd6bf8df3c8",
		Messages:       []Message{{Role: RoleUser, Content: []Block{{Type: BlockTypeText, Text: &Text{Content: "hi"}}}}},
	}, nil)
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	if res.LastSeq != 1 || res.TurnID == "" || res.ConversationID == "" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestClientStreamHTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "bad request")
	}))
	defer ts.Close()

	client := NewClient(ts.URL, nil)
	_, err := client.Stream(context.Background(), Request{
		ConversationID: "e7fdf7ec-fce5-4ca6-a572-bfd6bf8df3c8",
		Messages:       []Message{{Role: RoleUser, Content: []Block{{Type: BlockTypeText, Text: &Text{Content: "hi"}}}}},
	}, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var httpErr HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected HTTPError, got %T (%v)", err, err)
	}
}

func TestClientStream_InvalidContentType(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, nil)
	_, err := client.Stream(context.Background(), Request{
		ConversationID: "e7fdf7ec-fce5-4ca6-a572-bfd6bf8df3c8",
		Messages:       []Message{{Role: RoleUser, Content: []Block{{Type: BlockTypeText, Text: &Text{Content: "hi"}}}}},
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "expected text/event-stream") {
		t.Fatalf("expected content-type error, got %v", err)
	}
}

func TestClientStream_FallsBackToRequestConversationID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"chat_stream_version\":\"v2\",\"turn_id\":\"turn-1\",\"seq\":1,\"type\":\"text_delta\",\"text\":{\"content\":\"hello\"}}\n")
		_, _ = io.WriteString(w, "data: [DONE]\n")
	}))
	defer ts.Close()

	req := Request{
		ConversationID: "e7fdf7ec-fce5-4ca6-a572-bfd6bf8df3c8",
		Messages:       []Message{{Role: RoleUser, Content: []Block{{Type: BlockTypeText, Text: &Text{Content: "hi"}}}}},
	}
	client := NewClient(ts.URL, nil)
	res, err := client.Stream(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	if res.ConversationID != req.ConversationID {
		t.Fatalf("expected fallback conversation id %q, got %q", req.ConversationID, res.ConversationID)
	}
}
