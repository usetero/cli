package datadog

import (
	"errors"
	"testing"
)

func TestAppKeyErrorMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "nil error",
			err:  nil,
			want: "",
		},
		{
			name: "empty error",
			err:  errors.New(""),
			want: "Failed to connect Datadog. Please try again.",
		},
		{
			name: "strips graphql prefix",
			err:  errors.New("graphql: Invalid application key for site EU1"),
			want: "Invalid application key for site EU1",
		},
		{
			name: "maps network timeout",
			err:  errors.New("request timeout while contacting datadog"),
			want: "Could not reach Datadog. Check your connection and try again.",
		},
		{
			name: "passes through backend message",
			err:  errors.New("Application key must include logs_read_data permission"),
			want: "Application key must include logs_read_data permission",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := appKeyErrorMessage(tc.err)
			if got != tc.want {
				t.Fatalf("appKeyErrorMessage() = %q, want %q", got, tc.want)
			}
		})
	}
}
