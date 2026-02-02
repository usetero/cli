package api

import (
	"errors"
	"fmt"
	"testing"
)

func TestClassifyError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "nil error",
			err:  nil,
			want: nil,
		},
		{
			name: "not found",
			err:  errors.New("conversation not found"),
			want: ErrNotFound,
		},
		{
			name: "NOT_FOUND",
			err:  errors.New("ERROR: NOT_FOUND"),
			want: ErrNotFound,
		},
		{
			name: "not_found",
			err:  errors.New("resource_not_found"),
			want: ErrNotFound,
		},
		{
			name: "does not exist",
			err:  errors.New("conversation does not exist"),
			want: ErrNotFound,
		},
		{
			name: "already exists",
			err:  errors.New("conversation already exists"),
			want: ErrAlreadyExists,
		},
		{
			name: "ALREADY_EXISTS",
			err:  errors.New("ERROR: ALREADY_EXISTS"),
			want: ErrAlreadyExists,
		},
		{
			name: "duplicate",
			err:  errors.New("duplicate key violation"),
			want: ErrAlreadyExists,
		},
		{
			name: "conflict",
			err:  errors.New("conflict: resource exists"),
			want: ErrAlreadyExists,
		},
		{
			name: "unrelated error",
			err:  errors.New("network timeout"),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := classifyError(tt.err)
			// classifyError returns sentinel errors directly (ErrNotFound, ErrAlreadyExists, or nil)
			// so direct comparison is correct here
			if (got == nil) != (tt.want == nil) || (got != nil && tt.want != nil && !errors.Is(got, tt.want)) {
				t.Errorf("classifyError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClassifiedErrorsWorkWithErrorsIs(t *testing.T) {
	t.Parallel()

	t.Run("wrapped ErrNotFound", func(t *testing.T) {
		t.Parallel()
		original := errors.New("conversation not found")
		classified := classifyError(original)
		wrapped := fmt.Errorf("delete conversation abc: %w", classified)

		if !errors.Is(wrapped, ErrNotFound) {
			t.Error("errors.Is(wrapped, ErrNotFound) = false, want true")
		}
	})

	t.Run("wrapped ErrAlreadyExists", func(t *testing.T) {
		t.Parallel()
		original := errors.New("conversation already exists")
		classified := classifyError(original)
		wrapped := fmt.Errorf("create conversation abc: %w", classified)

		if !errors.Is(wrapped, ErrAlreadyExists) {
			t.Error("errors.Is(wrapped, ErrAlreadyExists) = false, want true")
		}
	})

	t.Run("errors.Join preserves both errors", func(t *testing.T) {
		t.Parallel()
		contextErr := fmt.Errorf("delete conversation %s", "abc")
		joined := errors.Join(contextErr, ErrNotFound)

		if !errors.Is(joined, ErrNotFound) {
			t.Error("errors.Is(joined, ErrNotFound) = false, want true")
		}
	})
}
