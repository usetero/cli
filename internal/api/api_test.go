package api

import (
	"errors"
	"testing"
)

func TestIsNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "sentinel ErrNotFound",
			err:  ErrNotFound,
			want: true,
		},
		{
			name: "wrapped ErrNotFound",
			err:  errors.New("wrapped: " + ErrNotFound.Error()),
			want: true,
		},
		{
			name: "message contains 'not found'",
			err:  errors.New("conversation not found"),
			want: true,
		},
		{
			name: "message contains 'NOT_FOUND'",
			err:  errors.New("ERROR: NOT_FOUND"),
			want: true,
		},
		{
			name: "message contains 'not_found'",
			err:  errors.New("resource_not_found"),
			want: true,
		},
		{
			name: "unrelated error",
			err:  errors.New("network timeout"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAlreadyExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "sentinel ErrAlreadyExists",
			err:  ErrAlreadyExists,
			want: true,
		},
		{
			name: "message contains 'already exists'",
			err:  errors.New("conversation already exists"),
			want: true,
		},
		{
			name: "message contains 'ALREADY_EXISTS'",
			err:  errors.New("ERROR: ALREADY_EXISTS"),
			want: true,
		},
		{
			name: "message contains 'duplicate'",
			err:  errors.New("duplicate key violation"),
			want: true,
		},
		{
			name: "message contains 'conflict'",
			err:  errors.New("conflict: resource exists"),
			want: true,
		},
		{
			name: "unrelated error",
			err:  errors.New("network timeout"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsAlreadyExists(tt.err); got != tt.want {
				t.Errorf("IsAlreadyExists() = %v, want %v", got, tt.want)
			}
		})
	}
}
