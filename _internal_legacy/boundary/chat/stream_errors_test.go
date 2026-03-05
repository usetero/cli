package chat

import (
	"context"
	"errors"
	"testing"
)

func TestClassifyStreamError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want StreamErrorClass
	}{
		{name: "context canceled", err: context.Canceled, want: StreamErrorClassCancelled},
		{name: "deadline exceeded", err: context.DeadlineExceeded, want: StreamErrorClassTimeout},
		{name: "user cancelled text", err: errors.New("user_cancelled"), want: StreamErrorClassCancelled},
		{name: "protocol", err: errors.New("protocol error: unknown event type"), want: StreamErrorClassProtocol},
		{name: "parse event", err: errors.New("parse event: bad json"), want: StreamErrorClassProtocol},
		{name: "server", err: errors.New("server error: internal error"), want: StreamErrorClassServer},
		{name: "server 400", err: errors.New("server error: stream: POST https://x: 400 Bad Request"), want: StreamErrorClassRequest},
		{name: "chat api", err: errors.New("chat API error 500"), want: StreamErrorClassServer},
		{name: "chat api 400", err: errors.New("chat API error 400: bad request"), want: StreamErrorClassRequest},
		{name: "unknown", err: errors.New("boom"), want: StreamErrorClassUnknown},
		{name: "nil", err: nil, want: StreamErrorClassUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ClassifyStreamError(tt.err); got != tt.want {
				t.Fatalf("ClassifyStreamError() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUserFacingStreamError(t *testing.T) {
	t.Parallel()

	if got := UserFacingStreamError(errors.New("chat API error 400: bad request")); got != "The request was rejected by the chat service. Please retry." {
		t.Fatalf("400 message = %q", got)
	}
	if got := UserFacingStreamError(errors.New("chat API error 500: internal")); got != "The chat service returned an internal error. Please try again." {
		t.Fatalf("500 message = %q", got)
	}
}
