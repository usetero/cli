package usecase

import (
	"context"
	"testing"
)

func TestChatBoundaryStreamErrorMapper(t *testing.T) {
	t.Parallel()

	m := NewChatBoundaryStreamErrorMapper()
	if got := m.Classify(context.Canceled); got == "" {
		t.Fatal("classify should not be empty")
	}
	if got := m.UserFacing(context.Canceled); got == "" {
		t.Fatal("user-facing should not be empty")
	}
}
