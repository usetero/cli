package auth

import (
	"testing"
	"time"
)

func TestAuthPollInterval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		interval int
		want     time.Duration
	}{
		{name: "zero uses default", interval: 0, want: defaultPollInterval},
		{name: "negative uses default", interval: -1, want: defaultPollInterval},
		{name: "sub-second clamps to minimum", interval: 1, want: minPollInterval},
		{name: "provider interval under cap", interval: 2, want: 2 * time.Second},
		{name: "provider interval over cap", interval: 5, want: maxPollInterval},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := authPollInterval(tc.interval)
			if got != tc.want {
				t.Fatalf("authPollInterval(%d) = %s, want %s", tc.interval, got, tc.want)
			}
		})
	}
}
