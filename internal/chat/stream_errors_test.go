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
		{name: "chat api", err: errors.New("chat API error 500"), want: StreamErrorClassServer},
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

	if got := UserFacingStreamError(errors.New("server error: internal error")); got == "" {
		t.Fatal("expected non-empty message")
	}
}
