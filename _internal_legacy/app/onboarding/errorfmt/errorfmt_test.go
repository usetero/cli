package errorfmt

import (
	"errors"
	"testing"
)

func TestUserFacing(t *testing.T) {
	t.Parallel()

	joined := errors.Join(errors.New("graphql: control plane rejected credentials"), errors.New("secondary"))

	tests := []struct {
		name     string
		err      error
		fallback string
		want     string
	}{
		{
			name:     "nil error uses fallback",
			err:      nil,
			fallback: "fallback",
			want:     "fallback",
		},
		{
			name:     "empty error uses fallback",
			err:      errors.New(""),
			fallback: "fallback",
			want:     "fallback",
		},
		{
			name:     "strips graphql prefix",
			err:      errors.New("graphql: invalid API key"),
			fallback: "fallback",
			want:     "invalid API key",
		},
		{
			name:     "handles joined errors",
			err:      joined,
			fallback: "fallback",
			want:     "control plane rejected credentials",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := UserFacing(tc.err, tc.fallback)
			if got != tc.want {
				t.Fatalf("UserFacing() = %q, want %q", got, tc.want)
			}
		})
	}
}
