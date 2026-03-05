package usecase

import (
	"errors"
	"testing"
)

type fakeErrorMapper struct {
	classify   string
	userFacing string
}

func (f fakeErrorMapper) Classify(error) string   { return f.classify }
func (f fakeErrorMapper) UserFacing(error) string { return f.userFacing }

func TestClassifyStreamError(t *testing.T) {
	t.Parallel()

	if got := ClassifyStreamError(nil, errors.New("x")); got != "unknown" {
		t.Fatalf("nil mapper classify = %q, want unknown", got)
	}
	if got := ClassifyStreamError(fakeErrorMapper{classify: "timeout"}, errors.New("x")); got != "timeout" {
		t.Fatalf("classify = %q, want timeout", got)
	}
}

func TestUserFacingStreamError(t *testing.T) {
	t.Parallel()

	if got := UserFacingStreamError(nil, errors.New("x")); got == "" {
		t.Fatal("nil mapper user-facing should not be empty")
	}
	if got := UserFacingStreamError(fakeErrorMapper{userFacing: "Readable"}, errors.New("x")); got != "Readable" {
		t.Fatalf("user-facing = %q, want Readable", got)
	}
}
